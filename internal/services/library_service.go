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

type LibraryService interface {
	RegisterLibrary(ctx context.Context, req dto.CreateLibraryRequest) (*models.Library, error)
	GetLibraryByID(ctx context.Context, id int) (*models.Library, error)
	GetLibraryLoans(ctx context.Context, id int) ([]models.Loan, error)
	GetLibraryBooks(ctx context.Context, id int) ([]models.Book, error)
	GetLibraryBooksByGenre(ctx context.Context, libraryID, genreID int) ([]models.Book, error)
	ListLibraries(ctx context.Context) ([]models.Library, error)
	UpdateLibrary(ctx context.Context, id int, req dto.CreateLibraryRequest) (*models.Library, error)
	DeleteLibrary(ctx context.Context, id int) error
}

type libraryService struct {
	db          *sql.DB
	libraryRepo repositories.LibraryRepository
	bookRepo    repositories.BookRepository
}

func NewLibraryService(db *sql.DB, libraryRepo repositories.LibraryRepository) LibraryService {
	return &libraryService{db: db, libraryRepo: libraryRepo}
}

// 1. RegisterLibrary
func (s *libraryService) RegisterLibrary(ctx context.Context, req dto.CreateLibraryRequest) (*models.Library, error) {
	library := &models.Library{
		Name:    req.Name,
		Address: req.Address,
	}

	if err := s.libraryRepo.Create(ctx, s.db, library); err != nil {
		return nil, fmt.Errorf("failed to register library: %w", err)
	}

	return library, nil
}

// 2. GetLibraryByID
func (s *libraryService) GetLibraryByID(ctx context.Context, id int) (*models.Library, error) {
	library, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error fetching library: %w", err)
	}
	if library == nil {
		return nil, errors.New("library branch record does not exist")
	}

	return library, nil
}

// 3. ListLibraries
func (s *libraryService) ListLibraries(ctx context.Context) ([]models.Library, error) {
	libraries, err := s.libraryRepo.ListAll(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list libraries: %w", err)
	}

	if libraries == nil {
		return []models.Library{}, nil // Prevent returning null to HTTP JSON responses
	}

	return libraries, nil
}

// 4. UpdateLibrary
func (s *libraryService) UpdateLibrary(ctx context.Context, id int, req dto.CreateLibraryRequest) (*models.Library, error) {
	// Verify library exists before updating
	existing, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if existing == nil {
		return nil, errors.New("cannot update: library branch record does not exist")
	}

	library := &models.Library{
		ID:      id,
		Name:    req.Name,
		Address: req.Address,
	}

	if err := s.libraryRepo.Update(ctx, s.db, library); err != nil {
		return nil, fmt.Errorf("failed to update library: %w", err)
	}

	return library, nil
}

// 5. DeleteLibrary
func (s *libraryService) DeleteLibrary(ctx context.Context, id int) error {
	// Verify library exists before attempting delete
	existing, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return fmt.Errorf("database error checking library before delete: %w", err)
	}
	if existing == nil {
		return errors.New("cannot delete: library branch record does not exist")
	}

	if err := s.libraryRepo.Delete(ctx, s.db, id); err != nil {
		return fmt.Errorf("failed to delete library branch: %w", err)
	}

	return nil
}

// 6. GetLibraryLoans
func (s *libraryService) GetLibraryLoans(ctx context.Context, id int) ([]models.Loan, error) {
	// 1. Verify library branch exists
	library, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if library == nil {
		return nil, errors.New("cannot fetch loans: targeted library branch does not exist")
	}

	// 2. Fetch loans originated at this branch
	loans, err := s.libraryRepo.ListAllLoans(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library loans: %w", err)
	}

	if loans == nil {
		return []models.Loan{}, nil
	}

	return loans, nil
}

// 7. GetLibraryBooks
func (s *libraryService) GetLibraryBooks(ctx context.Context, id int) ([]models.Book, error) {
	// 1. Verify library branch exists
	library, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if library == nil {
		return nil, errors.New("cannot fetch books: targeted library branch does not exist")
	}

	// 2. Fetch books stocked at this branch
	books, err := s.libraryRepo.ListAllBooks(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library books: %w", err)
	}

	if books == nil {
		return []models.Book{}, nil
	}

	return books, nil
}

func (s *libraryService) GetLibraryBooksByGenre(ctx context.Context, libraryID, genreID int) ([]models.Book, error) {
	// 1. Verify library exists
	if _, err := s.libraryRepo.GetByID(ctx, s.db, libraryID); err != nil {
		return nil, fmt.Errorf("library not found: %w", err)
	}

	// 2. Fetch filtered books from repository
	books, err := s.bookRepo.GetBooksByLibraryAndGenre(ctx, s.db, libraryID, genreID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve books for library %d and genre %d: %w", libraryID, genreID, err)
	}

	return books, nil
}
