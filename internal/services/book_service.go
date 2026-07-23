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

	BorrowBook(ctx context.Context, req dto.BorrowBookRequest) error
	ReturnBook(ctx context.Context, req dto.ReturnBookRequest) error
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
		GenreID: *req.GenreID,
	}

	if err := s.bookRepo.Create(ctx, s.db, book, *req.GenreID, req.AvailableCopies); err != nil {
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
	book.GenreID = *req.GenreID

	if err := s.bookRepo.Update(ctx, s.db, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id int) error {
	return s.bookRepo.Delete(ctx, s.db, id)
}

// --- 1. BorrowBook ---
func (s *bookService) BorrowBook(ctx context.Context, req dto.BorrowBookRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Verify Member Exists
	member, err := s.memberRepo.GetByID(ctx, tx, req.MemberID)
	if err != nil {
		return fmt.Errorf("database error checking member: %w", err)
	}
	if member == nil {
		return errors.New("checkout blocked: targeted member record does not exist")
	}

	// 2. Verify Book Exists in Catalog
	book, err := s.bookRepo.GetByID(ctx, tx, req.BookID)
	if err != nil {
		return fmt.Errorf("database error checking book catalog: %w", err)
	}
	if book == nil {
		return errors.New("checkout blocked: targeted book record does not exist")
	}

	// 3. Verify Stock at Selected Library Branch
	availableCopies, err := s.bookInventoryRepo.GetAvailableCopies(ctx, tx, req.BookID, req.Borrowed_library_id)
	if err != nil {
		return fmt.Errorf("database error checking available copies: %w", err)
	}
	if availableCopies == nil {
		return errors.New("checkout blocked: this book is not stocked at the selected library branch")
	}
	if *availableCopies <= 0 {
		return errors.New("checkout blocked: zero inventory balance remains for this book at this branch")
	}
	// 4. Verify Active Loan
	hasLoan, err := s.bookRepo.HasActiveLoan(ctx, tx, req.BookID, req.MemberID)
	if err != nil {
		return fmt.Errorf("database error checking active loans: %w", err)
	}
	if hasLoan {
		return errors.New("this member has already borrowed this book and has not returned it yet")
	}

	// 5. Decrement Stock in `book_inventory`
	if err := s.bookInventoryRepo.DecrementInventory(ctx, tx, req.BookID, req.Borrowed_library_id); err != nil {
		return fmt.Errorf("failed to decrement inventory: %w", err)
	}

	// 6. Register Loan Record
	if err := s.bookRepo.CreateLoan(ctx, tx, req.BookID, req.MemberID, req.Borrowed_library_id); err != nil {
		return fmt.Errorf("failed to register loan log: %w", err)
	}

	return tx.Commit()
}

// --- 2. ReturnBook ---
func (s *bookService) ReturnBook(ctx context.Context, req dto.ReturnBookRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Verify Book Exists
	book, err := s.bookRepo.GetByID(ctx, tx, req.BookID)
	if err != nil {
		return fmt.Errorf("database error checking book catalog: %w", err)
	}
	if book == nil {
		return errors.New("return blocked: target catalog book no longer exists")
	}

	// 2. Verify Member Has an Active Loan to Return
	hasLoan, err := s.bookRepo.HasActiveLoan(ctx, tx, req.BookID, req.MemberID)
	if err != nil {
		return fmt.Errorf("database error checking active loan: %w", err)
	}
	if !hasLoan {
		return errors.New("return blocked: no active loan found for this member and book")
	}

	// 3. Mark Loan as Returned
	if err := s.bookRepo.UpdateLoanReturn(ctx, tx, req.BookID, req.MemberID, req.Returned_library_id); err != nil {
		return fmt.Errorf("failed to update loan record: %w", err)
	}

	// 4. Increment Stock in `book_inventory` for the Receiving Library Branch
	// Uses `CreateOrUpdate` / `IncrementInventory` so returning to a new branch works seamlessly
	if err := s.bookInventoryRepo.IncrementInventory(ctx, tx, req.BookID, req.Returned_library_id); err != nil {
		return fmt.Errorf("failed to increment inventory: %w", err)
	}

	return tx.Commit()
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
