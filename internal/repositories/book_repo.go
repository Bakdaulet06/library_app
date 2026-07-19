package repositories

import (
	"context"
	"database/sql"
	"errors"
	"library/internal/models"
	"time"
)

// DBTX acts as a shared interface accepting both *sql.DB and *sql.Tx.
// This allows us to run repository methods inside or outside transactions.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type BookRepository interface {
	Create(ctx context.Context, db DBTX, book *models.Book) error
	GetByID(ctx context.Context, db DBTX, id int) (*models.Book, error)
	GetByISBN(ctx context.Context, db DBTX, isbn string) (*models.Book, error)
	UpdateCopies(ctx context.Context, db DBTX, bookID int, change int) error
	ListAvailable(ctx context.Context, db DBTX) ([]models.Book, error)

	// Loan sub-domain interactions
	CreateLoan(ctx context.Context, db DBTX, loan *models.Loan) error
	FindActiveLoan(ctx context.Context, db DBTX, bookID, memberID int) (*models.Loan, error)
	UpdateLoanReturn(ctx context.Context, db DBTX, loanID int, returnTime time.Time) error
}

type bookRepository struct{}

func NewBookRepository() BookRepository {
	return &bookRepository{}
}

func (r *bookRepository) Create(ctx context.Context, db DBTX, b *models.Book) error {
	query := `
		INSERT INTO books (title, author, isbn, available_copies)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;`

	err := db.QueryRowContext(ctx, query, b.Title, b.Author, b.ISBN, b.AvailableCopies).
		Scan(&b.ID, &b.CreatedAt)
	return err
}

func (r *bookRepository) GetByID(ctx context.Context, db DBTX, id int) (*models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books WHERE id = $1;`
	var b models.Book
	err := db.QueryRowContext(ctx, query, id).
		Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.AvailableCopies, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *bookRepository) GetByISBN(ctx context.Context, db DBTX, isbn string) (*models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books WHERE isbn = $1;`
	var b models.Book
	err := db.QueryRowContext(ctx, query, isbn).
		Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.AvailableCopies, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *bookRepository) UpdateCopies(ctx context.Context, db DBTX, bookID int, change int) error {
	// Using PostgreSQL math additions safely inside the query string
	query := `UPDATE books SET available_copies = available_copies + $1 WHERE id = $2;`
	res, err := db.ExecContext(ctx, query, change, bookID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("book record not found to update copies")
	}
	return nil
}

func (r *bookRepository) ListAvailable(ctx context.Context, db DBTX) ([]models.Book, error) {
	query := `SELECT id, title, author, isbn, available_copies, created_at FROM books WHERE available_copies > 0 ORDER BY id DESC;`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.AvailableCopies, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *bookRepository) CreateLoan(ctx context.Context, db DBTX, l *models.Loan) error {
	query := `
		INSERT INTO loans (book_id, member_id)
		VALUES ($1, $2)
		RETURNING id, loan_date;`
	err := db.QueryRowContext(ctx, query, l.BookID, l.MemberID).Scan(&l.ID, &l.LoanDate)
	return err
}

func (r *bookRepository) FindActiveLoan(ctx context.Context, db DBTX, bookID, memberID int) (*models.Loan, error) {
	query := `SELECT id, book_id, member_id, loan_date, return_date FROM loans WHERE book_id = $1 AND member_id = $2 AND return_date IS NULL LIMIT 1;`
	var l models.Loan
	err := db.QueryRowContext(ctx, query, bookID, memberID).
		Scan(&l.ID, &l.BookID, &l.MemberID, &l.LoanDate, &l.ReturnDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &l, err
}

func (r *bookRepository) UpdateLoanReturn(ctx context.Context, db DBTX, loanID int, returnTime time.Time) error {
	query := `UPDATE loans SET return_date = $1 WHERE id = $2;`
	res, err := db.ExecContext(ctx, query, returnTime, loanID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("loan record not found or already returned")
	}
	return nil
}
