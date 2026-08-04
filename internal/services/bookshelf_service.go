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

type BookshelfService interface {
	CreateBookshelf(ctx context.Context, shelf *models.Bookshelf) error
	GetBookshelfByID(ctx context.Context, libraryID, shelfID int) (*models.Bookshelf, error)
	GetBookshelvesByLibraryID(ctx context.Context, p params.BookshelfParams) ([]models.Bookshelf, error)
	GetBooksByShelfID(ctx context.Context, libraryID, shelfID int, p params.Pagination) ([]dto.BookWithShelfStockResponse, error)
	GetBookByShelfID(ctx context.Context, libraryID, shelfID, bookID int) (*dto.BookWithShelfStockResponse, error)
	DeleteBookshelf(ctx context.Context, libraryID, shelfID int) error
}

type bookshelfService struct {
	db            *sql.DB
	bookshelfRepo repositories.BookshelfRepository
	libraryRepo   repositories.LibraryRepository
	inventoryRepo repositories.BookInventoryRepository
}

func NewBookshelfService(
	db *sql.DB,
	bookshelfRepo repositories.BookshelfRepository,
	libraryRepo repositories.LibraryRepository,
	inventoryRepo repositories.BookInventoryRepository,
) BookshelfService {
	return &bookshelfService{
		db:            db,
		bookshelfRepo: bookshelfRepo,
		libraryRepo:   libraryRepo,
		inventoryRepo: inventoryRepo,
	}
}

// CreateBookshelf handles validation and creates a new shelf inside a library
func (s *bookshelfService) CreateBookshelf(ctx context.Context, shelf *models.Bookshelf) error {
	// 1. Verify target library branch exists
	if _, err := s.libraryRepo.GetByID(ctx, s.db, shelf.LibraryID); err != nil {
		return fmt.Errorf("library branch not found: %w", err)
	}

	code, err := s.bookshelfRepo.GetNextBookshelfCode(ctx, s.db)
	if err != nil {
		return err
	}
	// 2. Create the shelf (Repository sets initial empty_space = capacity)
	return s.bookshelfRepo.Create(ctx, s.db, shelf, code)
}

// GetBookshelfByID retrieves a single bookshelf
func (s *bookshelfService) GetBookshelfByID(ctx context.Context, libraryID, shelfID int) (*models.Bookshelf, error) {
	if shelfID <= 0 {
		return nil, errors.New("invalid bookshelf id")
	}
	return s.bookshelfRepo.GetByID(ctx, s.db, libraryID, shelfID)
}

// GetBookshelvesByLibraryID lists all bookshelves belonging to a library branch
func (s *bookshelfService) GetBookshelvesByLibraryID(ctx context.Context, p params.BookshelfParams) ([]models.Bookshelf, error) {
	// 1. Check if the library exists
	exists, err := s.libraryRepo.LibraryExistsByID(ctx, s.db, p.LibraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check library existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("Library with id %d doesn't exist", p.LibraryID) // Returns 404 domain error
	}

	// 2. Fetch bookshelves
	return s.bookshelfRepo.GetByLibraryID(ctx, s.db, p)
}

func (s *bookshelfService) GetBooksByShelfID(ctx context.Context, libraryID, shelfID int, p params.Pagination) ([]dto.BookWithShelfStockResponse, error) {
	exists, err := s.inventoryRepo.ShelfAndLibraryExists(ctx, s.db, libraryID, shelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to check inventory existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("Invalid bookshelf and library id parameters") // Returns 404 domain error
	}
	books, err := s.bookshelfRepo.GetBooksByShelfID(ctx, s.db, libraryID, shelfID, p)
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (s *bookshelfService) GetBookByShelfID(ctx context.Context, libraryID, shelfID, bookID int) (*dto.BookWithShelfStockResponse, error) {
	book, err := s.bookshelfRepo.GetBookByShelfID(ctx, s.db, libraryID, shelfID, bookID)
	if err != nil {
		return nil, err
	}
	return book, nil
}

// DeleteBookshelf removes a shelf by ID
func (s *bookshelfService) DeleteBookshelf(ctx context.Context, libraryID, shelfID int) error {
	if shelfID <= 0 {
		return errors.New("invalid bookshelf id")
	}
	return s.bookshelfRepo.Delete(ctx, s.db, libraryID, shelfID)
}
