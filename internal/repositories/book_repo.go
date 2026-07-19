package repositories

import (
	"context"
	"database/sql"
	"errors"
	"library/internal/models"
)

// GormExecutor standardizes query methods to support both standard *sql.DB pools and active *sql.Tx transactions transparently.
type GormExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type BookRepository interface {
	Create(ctx context.Context, exec GormExecutor, b *models.Book) error
	GetByID(ctx context.Context, exec GormExecutor, id int64) (*models.Book, error)
	GetByISBN(ctx context.Context, exec GormExecutor, isbn string) (*models.Book, error)
	ListAvailable(ctx context.Context, exec GormExecutor) ([]models.Book, error)
	ListAll(ctx context.Context, exec GormExecutor) ([]models.Book, error)
	Update(ctx context.Context, exec GormExecutor, b *models.Book) error
	Delete(ctx context.Context, exec GormExecutor, id int64) error

	// Transactional & Borrowing State Assertions
	HasActiveLoan(ctx context.Context, exec GormExecutor, bookID, memberID int64) (bool, error)
	CreateLoan(ctx context.Context, exec GormExecutor, bookID, memberID int64) error
	UpdateLoanReturn(ctx context.Context, exec GormExecutor, bookID, memberID int64) error
	ListLoans(ctx context.Context, exec GormExecutor) ([]models.Loan, error)
}

type bookRepository struct{}

func NewBookRepository() BookRepository {
	return &bookRepository{}
}

func (r *bookRepository) Create(ctx context.Context, exec GormExecutor, b *models.Book) error {
	query := `INSERT INTO books (title, author, isbn, available_copies) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return exec.QueryRowContext(ctx, query, b.Title, b.Author, b.Isbn, b.AvailableCopies).Scan(&b.ID, &b.CreatedAt)
}

func (r *bookRepository) GetByID(ctx context.Context, exec GormExecutor, id int64) (*models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books WHERE id = $1`
	var b models.Book
	err := exec.QueryRowContext(ctx, query, id).Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.AvailableCopies, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *bookRepository) GetByISBN(ctx context.Context, exec GormExecutor, isbn string) (*models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books WHERE isbn = $1`
	var b models.Book
	err := exec.QueryRowContext(ctx, query, isbn).Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.AvailableCopies, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *bookRepository) ListAvailable(ctx context.Context, exec GormExecutor) ([]models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books WHERE available_copies > 0 ORDER BY id ASC`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.AvailableCopies, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *bookRepository) ListAll(ctx context.Context, exec GormExecutor) ([]models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books ORDER BY id ASC`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.AvailableCopies, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *bookRepository) Update(ctx context.Context, exec GormExecutor, b *models.Book) error {
	query := `UPDATE books SET title = $1, author = $2, isbn = $3, available_copies = $4 WHERE id = $5`
	res, err := exec.ExecContext(ctx, query, b.Title, b.Author, b.Isbn, b.AvailableCopies, b.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("book record not found for update operations")
	}
	return err
}

func (r *bookRepository) Delete(ctx context.Context, exec GormExecutor, id int64) error {
	query := `DELETE FROM books WHERE id = $1`
	res, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("book record not found for deletion operations")
	}
	return err
}

func (r *bookRepository) HasActiveLoan(ctx context.Context, exec GormExecutor, bookID, memberID int64) (bool, error) {
	// Updated to use your table's return_date column name
	query := `SELECT EXISTS(SELECT 1 FROM loans WHERE book_id = $1 AND member_id = $2 AND return_date IS NULL)`
	var exists bool
	err := exec.QueryRowContext(ctx, query, bookID, memberID).Scan(&exists)
	return exists, err
}

func (r *bookRepository) CreateLoan(ctx context.Context, exec GormExecutor, bookID, memberID int64) error {
	// Updated to use your table's loan_date column name
	query := `INSERT INTO loans (book_id, member_id, loan_date) VALUES ($1, $2, NOW())`
	_, err := exec.ExecContext(ctx, query, bookID, memberID)
	return err
}

func (r *bookRepository) UpdateLoanReturn(ctx context.Context, exec GormExecutor, bookID, memberID int64) error {
	// Updated to use your table's return_date column name
	query := `UPDATE loans SET return_date = NOW() WHERE book_id = $1 AND member_id = $2 AND return_date IS NULL`
	res, err := exec.ExecContext(ctx, query, bookID, memberID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("no matching active loan record found for this member and book combination")
	}
	return err
}

func (r *bookRepository) ListLoans(ctx context.Context, exec GormExecutor) ([]models.Loan, error) {
	// Matching your precise database schema column mappings
	query := `SELECT id, book_id, member_id, loan_date, return_date FROM loans ORDER BY id DESC`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		if err := rows.Scan(&l.ID, &l.BookID, &l.MemberID, &l.BorrowedAt, &l.ReturnedAt); err != nil {
			return nil, err
		}
		loans = append(loans, l)
	}
	return loans, nil
}
