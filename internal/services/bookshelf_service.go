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

type BookshelfService interface {
	CreateBookshelf(ctx context.Context, shelf *models.Bookshelf) error
	GetBookshelfByID(ctx context.Context, libraryID, shelfID int) (*models.Bookshelf, error)
	GetBookshelvesByLibraryID(ctx context.Context, libraryID int) ([]models.Bookshelf, error)
	GetBooksByShelfID(ctx context.Context, libraryID, shelfID int) ([]dto.BookWithShelfStockResponse, error)
	GetBookByShelfID(ctx context.Context, libraryID, shelfID, bookID int) (*dto.BookWithShelfStockResponse, error)
	DeleteBookshelf(ctx context.Context, libraryID, shelfID int) error
}

type bookshelfService struct {
	db            *sql.DB
	bookshelfRepo repositories.BookshelfRepository
	libraryRepo   repositories.LibraryRepository
}

func NewBookshelfService(
	db *sql.DB,
	bookshelfRepo repositories.BookshelfRepository,
	libraryRepo repositories.LibraryRepository,
) BookshelfService {
	return &bookshelfService{
		db:            db,
		bookshelfRepo: bookshelfRepo,
		libraryRepo:   libraryRepo,
	}
}

// CreateBookshelf handles validation and creates a new shelf inside a library
func (s *bookshelfService) CreateBookshelf(ctx context.Context, shelf *models.Bookshelf) error {
	// 1. Verify target library branch exists
	if _, err := s.libraryRepo.GetByID(ctx, s.db, shelf.LibraryID); err != nil {
		return fmt.Errorf("library branch not found: %w", err)
	}

	// 2. Create the shelf (Repository sets initial empty_space = capacity)
	return s.bookshelfRepo.Create(ctx, s.db, shelf)
}

// GetBookshelfByID retrieves a single bookshelf
func (s *bookshelfService) GetBookshelfByID(ctx context.Context, libraryID, shelfID int) (*models.Bookshelf, error) {
	if shelfID <= 0 {
		return nil, errors.New("invalid bookshelf id")
	}
	return s.bookshelfRepo.GetByID(ctx, s.db, libraryID, shelfID)
}

// GetBookshelvesByLibraryID lists all bookshelves belonging to a library branch
func (s *bookshelfService) GetBookshelvesByLibraryID(ctx context.Context, libraryID int) ([]models.Bookshelf, error) {
	if libraryID <= 0 {
		return nil, errors.New("invalid library id")
	}
	return s.bookshelfRepo.GetByLibraryID(ctx, s.db, libraryID)
}

func (s *bookshelfService) GetBooksByShelfID(ctx context.Context, libraryID, shelfID int) ([]dto.BookWithShelfStockResponse, error) {
	books, err := s.bookshelfRepo.GetBooksByShelfID(ctx, s.db, libraryID, shelfID)
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
