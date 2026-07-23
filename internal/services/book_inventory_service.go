package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"
)

type BookInventoryService interface {
	CreateOrUpdateBookInventory(ctx context.Context, req dto.CreateBookInventoryRequest) (*models.BookInventory, error)
	GetAvailableCopies(ctx context.Context, book_id, library_Id int) (*int, error)
	ListBookInventory(ctx context.Context) ([]models.BookInventory, error)
	DeleteBookInventory(ctx context.Context, libraryId, bookId int) error
}

type bookInventoryService struct {
	db                *sql.DB
	bookInventoryRepo repositories.BookInventoryRepository // Handles book_inventory table
	bookRepo          repositories.BookRepository          // Used to verify BookID exists
	libraryRepo       repositories.LibraryRepository
}

func NewBookInventoryService(
	db *sql.DB,
	bookInventoryRepo repositories.BookInventoryRepository,
	bookRepo repositories.BookRepository,
	libraryRepo repositories.LibraryRepository,
) BookInventoryService {
	return &bookInventoryService{
		db:                db,
		bookInventoryRepo: bookInventoryRepo,
		bookRepo:          bookRepo,
		libraryRepo:       libraryRepo,
	}
}

func (s *bookInventoryService) CreateOrUpdateBookInventory(ctx context.Context, req dto.CreateBookInventoryRequest) (*models.BookInventory, error) {
	// 1. Check if Book exists in catalog
	book, err := s.bookRepo.GetByID(ctx, s.db, req.BookID)
	if err != nil {
		return nil, fmt.Errorf("database error checking book catalog: %w", err)
	}
	if book == nil {
		return nil, errors.New("cannot add stock: targeted book record does not exist")
	}

	// 2. Check if Library exists
	library, err := s.libraryRepo.GetByID(ctx, s.db, req.LibraryID)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if library == nil {
		return nil, errors.New("cannot add stock: targeted library branch does not exist")
	}

	// 3. Prepare inventory model
	inventory := &models.BookInventory{
		LibraryID:       req.LibraryID,
		BookID:          req.BookID,
		AvailableCopies: req.AvailableCopies,
	}

	// 4. Save stock via bookInventoryRepo
	if err := s.bookInventoryRepo.CreateOrUpdate(ctx, s.db, inventory); err != nil {
		return nil, fmt.Errorf("failed to insert book inventory: %w", err)
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

// --- 2. ListBookInventory ---
func (s *bookInventoryService) ListBookInventory(ctx context.Context) ([]models.BookInventory, error) {
	inventory, err := s.bookInventoryRepo.List(ctx, s.db)
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
	// Ensure the stock record actually exists before attempting deletion
	existing, err := s.bookInventoryRepo.GetAvailableCopies(ctx, s.db, bookID, libraryID)
	if err != nil {
		return fmt.Errorf("database error checking inventory before delete: %w", err)
	}
	if existing == nil {
		return errors.New("cannot delete inventory: record does not exist for this library and book")
	}

	if err := s.bookInventoryRepo.Delete(ctx, s.db, libraryID, bookID); err != nil {
		return fmt.Errorf("failed to delete inventory record: %w", err)
	}

	return nil
}
