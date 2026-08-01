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

type BookService interface {
	CreateBook(ctx context.Context, req dto.CreateBookRequest) (*models.Book, error)
	GetBookByID(ctx context.Context, id int) (*models.Book, error)
	GetBooksByGenreID(ctx context.Context, genreID int) ([]models.Book, error)
	ListAvailableBooks(ctx context.Context) ([]models.Book, error)
	ListAllBooks(ctx context.Context) ([]models.Book, error)
	UpdateBook(ctx context.Context, id int, req dto.CreateBookRequest) (*models.Book, error)
	DeleteBook(ctx context.Context, id int) error

	ListAllLoans(ctx context.Context) ([]models.Loan, error)
}

type bookService struct {
	db                *sql.DB
	bookRepo          repositories.BookRepository
	memberRepo        repositories.MemberRepository
	bookInventoryRepo repositories.BookInventoryRepository
}

func NewBookService(db *sql.DB, bRepo repositories.BookRepository, mRepo repositories.MemberRepository, biRepo repositories.BookInventoryRepository) BookService {
	return &bookService{db: db, bookRepo: bRepo, memberRepo: mRepo, bookInventoryRepo: biRepo}
}

func (s *bookService) CreateBook(ctx context.Context, req dto.CreateBookRequest) (*models.Book, error) {
	existing, err := s.bookRepo.GetByISBN(ctx, s.db, req.Isbn)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("a book with this isbn code already exists inside the inventory system")
	}

	book := &models.Book{
		Title:   req.Title,
		Author:  req.Author,
		Isbn:    req.Isbn,
		GenreID: req.GenreID,
		Price:   req.Price,
	}

	if err := s.bookRepo.Create(ctx, s.db, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) GetBookByID(ctx context.Context, id int) (*models.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.New("book record not found")
	}
	return book, nil
}

func (s *bookService) ListAvailableBooks(ctx context.Context) ([]models.Book, error) {
	return s.bookRepo.ListAvailable(ctx, s.db)
}

func (s *bookService) ListAllBooks(ctx context.Context) ([]models.Book, error) {
	return s.bookRepo.ListAll(ctx, s.db)
}

func (s *bookService) UpdateBook(ctx context.Context, id int, req dto.CreateBookRequest) (*models.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.New("target book record not found")
	}

	if req.Isbn != book.Isbn {
		duplicate, err := s.bookRepo.GetByISBN(ctx, s.db, req.Isbn)
		if err != nil {
			return nil, err
		}
		if duplicate != nil {
			return nil, errors.New("another book has already claimed this updated isbn code")
		}
	}

	book.Title = req.Title
	book.Author = req.Author
	book.Isbn = req.Isbn
	book.GenreID = req.GenreID
	book.Price = req.Price

	if err := s.bookRepo.Update(ctx, s.db, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id int) error {
	return s.bookRepo.Delete(ctx, s.db, id)
}

// --- 3. ListAllLoans ---
func (s *bookService) ListAllLoans(ctx context.Context) ([]models.Loan, error) {
	loans, err := s.bookRepo.ListLoans(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list loans: %w", err)
	}
	if loans == nil {
		return []models.Loan{}, nil // Return empty slice instead of nil
	}
	return loans, nil
}

func (s *bookService) GetBooksByGenreID(ctx context.Context, genreID int) ([]models.Book, error) {
	// Optional: You can check if the genre exists here using your genreRepo if you have one
	books, err := s.bookRepo.GetByGenreID(ctx, s.db, genreID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve books for genre ID %d: %w", genreID, err)
	}

	return books, nil
}
