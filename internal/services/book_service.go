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
	GetBookByID(ctx context.Context, id int64) (*models.Book, error)
	ListAvailableBooks(ctx context.Context) ([]models.Book, error)
	ListAllBooks(ctx context.Context) ([]models.Book, error)
	UpdateBook(ctx context.Context, id int64, req dto.CreateBookRequest) (*models.Book, error)
	DeleteBook(ctx context.Context, id int64) error

	BorrowBook(ctx context.Context, req dto.BorrowBookRequest) error
	ReturnBook(ctx context.Context, req dto.ReturnBookRequest) error
	ListAllLoans(ctx context.Context) ([]models.Loan, error)
}

type bookService struct {
	db         *sql.DB
	bookRepo   repositories.BookRepository
	memberRepo repositories.MemberRepository
}

func NewBookService(db *sql.DB, bRepo repositories.BookRepository, mRepo repositories.MemberRepository) BookService {
	return &bookService{db: db, bookRepo: bRepo, memberRepo: mRepo}
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
		Title:           req.Title,
		Author:          req.Author,
		Isbn:            req.Isbn,
		AvailableCopies: req.AvailableCopies,
	}

	if err := s.bookRepo.Create(ctx, s.db, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) GetBookByID(ctx context.Context, id int64) (*models.Book, error) {
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

func (s *bookService) UpdateBook(ctx context.Context, id int64, req dto.CreateBookRequest) (*models.Book, error) {
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
	book.AvailableCopies = req.AvailableCopies

	if err := s.bookRepo.Update(ctx, s.db, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id int64) error {
	return s.bookRepo.Delete(ctx, s.db, id)
}

func (s *bookService) BorrowBook(ctx context.Context, req dto.BorrowBookRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	member, err := s.memberRepo.GetByID(ctx, tx, req.MemberID)
	if err != nil || member == nil {
		return errors.New("checkout blocked: targeted member record does not exist")
	}

	book, err := s.bookRepo.GetByID(ctx, tx, req.BookID)
	if err != nil || book == nil {
		return errors.New("checkout blocked: targeted book record does not exist")
	}

	if book.AvailableCopies <= 0 {
		return errors.New("checkout blocked: zero inventory balance remains for this book")
	}

	hasLoan, err := s.bookRepo.HasActiveLoan(ctx, tx, req.BookID, req.MemberID)
	if err != nil {
		return err
	}
	if hasLoan {
		return errors.New("this member has already borrowed this book and has not returned it yet")
	}

	book.AvailableCopies--
	if err := s.bookRepo.Update(ctx, tx, book); err != nil {
		return fmt.Errorf("failed to decrement inventory: %w", err)
	}

	if err := s.bookRepo.CreateLoan(ctx, tx, req.BookID, req.MemberID); err != nil {
		return fmt.Errorf("failed to register loan log: %w", err)
	}

	return tx.Commit()
}

func (s *bookService) ReturnBook(ctx context.Context, req dto.ReturnBookRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	book, err := s.bookRepo.GetByID(ctx, tx, req.BookID)
	if err != nil || book == nil {
		return errors.New("return blocked: target catalog book no longer exists")
	}

	if err := s.bookRepo.UpdateLoanReturn(ctx, tx, req.BookID, req.MemberID); err != nil {
		return err
	}

	book.AvailableCopies++
	if err := s.bookRepo.Update(ctx, tx, book); err != nil {
		return fmt.Errorf("failed to increment inventory: %w", err)
	}

	return tx.Commit()
}

func (s *bookService) ListAllLoans(ctx context.Context) ([]models.Loan, error) {
	return s.bookRepo.ListLoans(ctx, s.db)
}
