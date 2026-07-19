package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"
)

type BookService interface {
	CreateBook(ctx context.Context, req dto.CreateBookRequest) (*models.Book, error)
	ListAvailableBooks(ctx context.Context) ([]models.Book, error)
	BorrowBook(ctx context.Context, req dto.BorrowBookRequest) error
	ReturnBook(ctx context.Context, req dto.ReturnBookRequest) error
}

type bookService struct {
	db         *sql.DB
	bookRepo   repositories.BookRepository
	memberRepo repositories.MemberRepository
}

func NewBookService(db *sql.DB, br repositories.BookRepository, mr repositories.MemberRepository) BookService {
	return &bookService{
		db:         db,
		bookRepo:   br,
		memberRepo: mr,
	}
}

func (s *bookService) CreateBook(ctx context.Context, req dto.CreateBookRequest) (*models.Book, error) {
	// Business Rule: Validate unique ISBN first
	existing, err := s.bookRepo.GetByISBN(ctx, s.db, req.ISBN)
	if err != nil {
		return nil, fmt.Errorf("error verifying unique isbn: %w", err)
	}
	if existing != nil {
		return nil, errors.New("a book with this isbn already exists")
	}

	book := &models.Book{
		Title:           req.Title,
		Author:          req.Author,
		ISBN:            req.ISBN,
		AvailableCopies: req.AvailableCopies,
	}

	if err := s.bookRepo.Create(ctx, s.db, book); err != nil {
		return nil, fmt.Errorf("failed to save book record: %w", err)
	}

	return book, nil
}

func (s *bookService) ListAvailableBooks(ctx context.Context) ([]models.Book, error) {
	books, err := s.bookRepo.ListAvailable(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve available books: %w", err)
	}
	return books, nil
}

func (s *bookService) BorrowBook(ctx context.Context, req dto.BorrowBookRequest) error {
	// Atomic transaction initiation
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defers structural clean up via rollback pattern
	defer tx.Rollback()

	// 1. Verify Member status explicitly
	member, err := s.memberRepo.GetByID(ctx, tx, req.MemberID)
	if err != nil {
		return fmt.Errorf("database query error fetching member: %w", err)
	}
	if member == nil {
		return errors.New("member profile does not exist")
	}

	// 2. Verify book profile exists and check stock limits
	book, err := s.bookRepo.GetByID(ctx, tx, req.BookID)
	if err != nil {
		return fmt.Errorf("database query error fetching book: %w", err)
	}
	if book == nil {
		return errors.New("book target profile does not exist")
	}
	if book.AvailableCopies <= 0 {
		return errors.New("no available copies left in library inventory")
	}

	// 3. Prevent duplicate active borrowings
	active, err := s.bookRepo.FindActiveLoan(ctx, tx, req.BookID, req.MemberID)
	if err != nil {
		return fmt.Errorf("database query checking loans: %w", err)
	}
	if active != nil {
		return errors.New("this member has already borrowed this book and has not returned it yet")
	}

	// 4. Record the checkout transaction
	loan := &models.Loan{
		BookID:   req.BookID,
		MemberID: req.MemberID,
	}
	if err := s.bookRepo.CreateLoan(ctx, tx, loan); err != nil {
		return fmt.Errorf("failed to log loan statement: %w", err)
	}

	// 5. Decrement inventory level by 1
	if err := s.bookRepo.UpdateCopies(ctx, tx, req.BookID, -1); err != nil {
		return fmt.Errorf("failed to decrement book inventory: %w", err)
	}

	return tx.Commit()
}

func (s *bookService) ReturnBook(ctx context.Context, req dto.ReturnBookRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Trace exact ongoing loan record
	loan, err := s.bookRepo.FindActiveLoan(ctx, tx, req.BookID, req.MemberID)
	if err != nil {
		return fmt.Errorf("database checking active loan record: %w", err)
	}
	if loan == nil {
		return errors.New("no active borrowing transaction found matching this member and book combination")
	}

	// 2. Set the return timeline status
	now := time.Now()
	if err := s.bookRepo.UpdateLoanReturn(ctx, tx, loan.ID, now); err != nil {
		return fmt.Errorf("failed to process loan closure: %w", err)
	}

	// 3. Increment inventory level by 1
	if err := s.bookRepo.UpdateCopies(ctx, tx, req.BookID, 1); err != nil {
		return fmt.Errorf("failed to increment book inventory: %w", err)
	}

	return tx.Commit()
}
