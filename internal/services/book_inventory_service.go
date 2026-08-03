package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/dto"
	"library/internal/models"
	"library/internal/params"
	"library/internal/repositories"
)

type BookInventoryService interface {
	AddBookInventory(ctx context.Context, req dto.CreateBookInventoryRequest) (*models.BookInventory, error)
	GetAvailableCopies(ctx context.Context, book_id, library_Id int) (*int, error)
	ListBookInventory(ctx context.Context, p params.Pagination) ([]models.BookInventory, error)
	DeleteBookInventory(ctx context.Context, libraryId, bookId int) error
}

type bookInventoryService struct {
	db                *sql.DB
	bookInventoryRepo repositories.BookInventoryRepository // Handles book_inventory table
	bookRepo          repositories.BookRepository          // Used to verify BookID exists
	libraryRepo       repositories.LibraryRepository
	bookshelfRepo     repositories.BookshelfRepository
}

func NewBookInventoryService(
	db *sql.DB,
	bookInventoryRepo repositories.BookInventoryRepository,
	bookRepo repositories.BookRepository,
	libraryRepo repositories.LibraryRepository,
	bookshelfRepo repositories.BookshelfRepository,
) BookInventoryService {
	return &bookInventoryService{
		db:                db,
		bookInventoryRepo: bookInventoryRepo,
		bookRepo:          bookRepo,
		libraryRepo:       libraryRepo,
		bookshelfRepo:     bookshelfRepo,
	}
}

func (s *bookInventoryService) AddBookInventory(
	ctx context.Context,
	req dto.CreateBookInventoryRequest,
) (*models.BookInventory, error) {

	// 1. Begin Database Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Clean defer: automatically rolls back if tx hasn't been committed yet
	defer tx.Rollback()

	// 2. Validate Book existence
	book, err := s.bookRepo.GetByID(ctx, tx, req.BookID)
	if err != nil || book == nil {
		return nil, fmt.Errorf("invalid book ID: %w", err)
	}

	// 3. Validate Library existence
	library, err := s.libraryRepo.GetByID(ctx, tx, req.LibraryID)
	if err != nil || library == nil {
		return nil, fmt.Errorf("invalid library ID: %w", err)
	}

	// 4. Validate Bookshelf existence
	shelf, err := s.bookshelfRepo.GetByID(ctx, tx, req.LibraryID, req.BookshelfID)
	if err != nil || shelf == nil {
		return nil, fmt.Errorf("invalid bookshelf ID: bookshelf does not exist")
	}

	// 5. Reserve shelf space for the incoming copies
	err = s.bookshelfRepo.UpdateEmptySpace(ctx, tx, req.LibraryID, req.BookshelfID, req.AvailableCopies)
	if err != nil {
		return nil, fmt.Errorf("shelf space reservation failed: %w", err)
	}

	// 6. Add copies to inventory
	inventory, err := s.bookInventoryRepo.AddCopies(ctx, tx, &models.BookInventory{
		BookLocation: models.BookLocation{
			LibraryID:   req.LibraryID,
			BookshelfID: req.BookshelfID,
			BookID:      req.BookID,
		},
		AvailableCopies: req.AvailableCopies,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update book inventory: %w", err)
	}

	// 7. Commit Transaction
	// FIX: tx.Commit() returns error directly; calling .Error is invalid
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return inventory, nil
}

func (s *bookInventoryService) GetAvailableCopies(ctx context.Context, bookID, libraryID int) (*int, error) {
	// First, verify the inventory record exists for this specific branch
	copies, err := s.bookInventoryRepo.GetAvailableCopies(ctx, s.db, bookID, libraryID)
	if err != nil {
		return nil, fmt.Errorf("database error fetching copies: %w", err)
	}

	if copies == nil {
		return nil, errors.New("this book is not stocked at the specified library branch")
	}

	return copies, nil
}

func (s *bookInventoryService) ListBookInventory(ctx context.Context, p params.Pagination) ([]models.BookInventory, error) {
	inventory, err := s.bookInventoryRepo.List(ctx, s.db, p)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch book inventory list: %w", err)
	}

	if inventory == nil {
		return []models.BookInventory{}, nil // Return empty slice instead of nil for clean JSON responses ([])
	}

	return inventory, nil
}

// --- 4. DeleteBookInventory ---
// Removes a book listing from a specific library's inventory
func (s *bookInventoryService) DeleteBookInventory(ctx context.Context, libraryID, bookID int) error {
	// 1. Begin Database Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // Safe rollback on error; no-op if Tx is already committed

	// 2. Retrieve all shelf assignments for this book in this library before deleting
	// Returns a slice of structs containing BookshelfID and Quantity
	shelfAllocations, err := s.bookInventoryRepo.GetShelfAllocationsByBook(ctx, tx, libraryID, bookID)
	if err != nil {
		return fmt.Errorf("failed to check book shelf locations: %w", err)
	}
	if len(shelfAllocations) == 0 {
		return errors.New("cannot delete inventory: record does not exist for this library and book")
	}

	// 3. Free up empty space on each bookshelf that stored this book
	for _, alloc := range shelfAllocations {
		// Passing positive quantity frees up empty_space on the bookshelf
		err := s.bookshelfRepo.UpdateEmptySpace(ctx, tx, libraryID, alloc.BookshelfID, -alloc.Quantity)
		if err != nil {
			return fmt.Errorf("failed to free bookshelf space for shelf %d: %w", alloc.BookshelfID, err)
		}
	}

	// 4. Delete inventory records for this book across this library
	if err := s.bookInventoryRepo.Delete(ctx, tx, libraryID, bookID); err != nil {
		return fmt.Errorf("failed to delete inventory records: %w", err)
	}

	// 5. Commit Transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
