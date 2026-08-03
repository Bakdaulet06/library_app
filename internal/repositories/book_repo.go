package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/models"
	"library/internal/params"
	"log"
	"strings"
)

// GormExecutor standardizes query methods to support both standard *sql.DB pools and active *sql.Tx transactions transparently.
type GormExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type BookRepository interface {
	Create(ctx context.Context, exec GormExecutor, b *models.Book) error
	GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Book, error)
	GetByISBN(ctx context.Context, exec GormExecutor, isbn string) (*models.Book, error)
	ListAll(ctx context.Context, exec GormExecutor, params params.BookParams) ([]models.Book, error)
	Update(ctx context.Context, exec GormExecutor, b *models.Book) error
	Delete(ctx context.Context, exec GormExecutor, id int) error

	// Transactional & Borrowing State Assertions
	HasActiveLoan(ctx context.Context, exec GormExecutor, bookID, memberID int) (bool, error)
	CreateLoan(ctx context.Context, exec GormExecutor, bookID, memberID, borrowed_library_id, borrowed_days int) error
	UpdateLoanReturn(ctx context.Context, exec GormExecutor, bookID, memberID, returned_library_id int) error
	ListLoans(ctx context.Context, exec GormExecutor) ([]models.Loan, error)
	GetLoansByMemberID(ctx context.Context, exec GormExecutor, memberID int) ([]models.Loan, error)

	CreateReturnedBook(ctx context.Context, exec GormExecutor, bookID, libraryID, memberID int) error
	GetReturnedBooksByLibrary(ctx context.Context, exec GormExecutor, libraryID int) ([]models.ReturnedBook, error)
	GetReturnedBook(ctx context.Context, exec GormExecutor, libraryID, bookID int) (*models.ReturnedBook, error)
	DeleteReturnedBook(ctx context.Context, exec GormExecutor, id int) error

	GetActiveLoan(ctx context.Context, exec GormExecutor, bookID, memberID int) (*models.Loan, error)
}

type bookRepository struct{}

func NewBookRepository() BookRepository {
	return &bookRepository{}
}

func (r *bookRepository) Create(ctx context.Context, exec GormExecutor, b *models.Book) error {
	// 1. Insert book
	queryBook := `INSERT INTO books (title, author, isbn, genre_id) 
                  VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := exec.QueryRowContext(ctx, queryBook, b.Title, b.Author, b.Isbn, b.GenreID).
		Scan(&b.ID, &b.CreatedAt)
	return err
}

func (r *bookRepository) GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Book, error) {
	query := `SELECT id, title, author, isbn, genre_id, created_at FROM books WHERE id = $1`
	var b models.Book
	err := exec.QueryRowContext(ctx, query, id).Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.GenreID, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *bookRepository) GetByISBN(ctx context.Context, exec GormExecutor, isbn string) (*models.Book, error) {
	query := `SELECT id, title, author, isbn, genre_id, created_at FROM books WHERE isbn = $1`
	var b models.Book
	err := exec.QueryRowContext(ctx, query, isbn).Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.GenreID, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *bookRepository) ListAll(ctx context.Context, exec GormExecutor, params params.BookParams) ([]models.Book, error) {
	query := `SELECT id, title, author, isbn, genre_id, price, created_at FROM books`

	var whereClauses []string
	var args []interface{}
	paramIdx := 1 // Postgres placeholder counter ($1, $2, etc.)

	// 1. Dynamic Search / Filtering
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		// Postgres uses $1, $2, $3...
		whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d OR author ILIKE $%d OR isbn ILIKE $%d)", paramIdx, paramIdx+1, paramIdx+2))
		args = append(args, searchTerm, searchTerm, searchTerm)
		paramIdx += 3
	}

	if params.GenreID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("genre_id = $%d", paramIdx))
		args = append(args, params.GenreID)
		paramIdx++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 2. Sorting Allowlist (Safe from SQL Injection)
	allowedColumns := map[string]string{
		"id":         "id",
		"title":      "title",
		"author":     "author",
		"created_at": "created_at",
		"price":      "price",
	}

	sortColumn, exists := allowedColumns[strings.ToLower(params.SortBy)]
	if !exists {
		sortColumn = "id"
	}

	order := strings.ToUpper(params.Order)
	if order != "DESC" {
		order = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, order)

	// 3. Limit & Offset with Postgres placeholders
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, params.Limit, params.Offset)

	// 4. Query Execution
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Isbn, &b.GenreID, &b.Price, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	return books, nil
}

func (r *bookRepository) Update(ctx context.Context, exec GormExecutor, b *models.Book) error {
	query := `UPDATE books SET title = $1, author = $2, isbn = $3, genre_id = $4, price = $5 WHERE id = $5`
	res, err := exec.ExecContext(ctx, query, b.Title, b.Author, b.Isbn, b.GenreID, b.Price, b.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("book record not found for update operations")
	}
	return err
}

func (r *bookRepository) Delete(ctx context.Context, exec GormExecutor, id int) error {
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

func (r *bookRepository) HasActiveLoan(ctx context.Context, exec GormExecutor, bookID, memberID int) (bool, error) {
	// Updated to use your table's return_date column name
	query := `SELECT EXISTS(SELECT 1 FROM loans WHERE book_id = $1 AND member_id = $2 AND returned_at IS NULL)`
	var exists bool
	err := exec.QueryRowContext(ctx, query, bookID, memberID).Scan(&exists)
	return exists, err
}

func (r *bookRepository) CreateLoan(ctx context.Context, exec GormExecutor, bookID, memberID, borrowedLibraryID, borrowedDays int) error {
	query := `
        INSERT INTO loans (book_id, member_id, borrowed_library_id, borrowed_at, borrowed_days) 
        VALUES ($1, $2, $3, NOW(), $4)`

	_, err := exec.ExecContext(ctx, query, bookID, memberID, borrowedLibraryID, borrowedDays)
	if err != nil {
		return fmt.Errorf("failed to insert loan record: %w", err)
	}
	return nil
}

func (r *bookRepository) UpdateLoanReturn(ctx context.Context, exec GormExecutor, bookID, memberID, returnedLibraryID int) error {
	query := `
        UPDATE loans 
        SET returned_at = NOW(), returned_library_id = $1 
        WHERE book_id = $2 AND member_id = $3 AND returned_at IS NULL`

	res, err := exec.ExecContext(ctx, query, returnedLibraryID, bookID, memberID)
	if err != nil {
		return fmt.Errorf("failed to execute loan return update: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no matching active loan record found for this member and book combination")
	}
	return nil
}

func (r *bookRepository) ListLoans(ctx context.Context, exec GormExecutor) ([]models.Loan, error) {
	query := `
		SELECT 
			id,
			book_id,
			member_id,
			borrowed_at,
			returned_at,
			borrowed_library_id,
			returned_library_id
		FROM loans
		ORDER BY id DESC
	`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		err := rows.Scan(
			&l.ID,
			&l.BookID,
			&l.MemberID,
			&l.BorrowedAt,
			&l.ReturnedAt, // *time.Time handles [null] nicely
			&l.BorrowedLibraryID,
			&l.ReturnedLibraryID, // *int handles [null] nicely!
		)
		fmt.Println(err)
		if err != nil {
			return nil, err
		}
		loans = append(loans, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}

func (r *bookRepository) GetLoansByMemberID(ctx context.Context, exec GormExecutor, memberID int) ([]models.Loan, error) {
	query := `
		SELECT id, book_id, member_id, borrowed_library_id, returned_library_id, 
		       borrowed_at, returned_at
		FROM loans
		WHERE member_id = $1
		ORDER BY borrowed_at DESC
	`

	rows, err := exec.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan

	for rows.Next() {
		var l models.Loan
		fmt.Printf("%T\n", l.ReturnedLibraryID)
		fmt.Printf("Type of ReturnedLibraryID: %T\n", l.ReturnedLibraryID)
		err := rows.Scan(
			&l.ID,
			&l.BookID,
			&l.MemberID,
			&l.BorrowedLibraryID,
			&l.ReturnedLibraryID,
			&l.BorrowedAt,
			&l.ReturnedAt, // database/sql handles null SQL timestamps directly into *time.Time
		)
		if err != nil {
			log.Printf("SCAN ERROR: %v\n", err)
			return nil, err
		}
		loans = append(loans, l)
	}

	return loans, nil
}

func (r *bookRepository) CreateReturnedBook(ctx context.Context, exec GormExecutor, bookID, libraryID, memberID int) error {
	query := `
		INSERT INTO returned_books (book_id, library_id, member_id)
		VALUES ($1, $2, $3);
	`
	_, err := exec.ExecContext(ctx, query, bookID, libraryID, memberID)
	if err != nil {
		return fmt.Errorf("failed to create returned book record: %w", err)
	}
	return nil
}

func (r *bookRepository) GetReturnedBooksByLibrary(ctx context.Context, exec GormExecutor, libraryID int) ([]models.ReturnedBook, error) {
	query := `
		SELECT id, book_id, library_id, member_id, returned_at
		FROM returned_books
		WHERE library_id = $1
		ORDER BY returned_at;
	`
	rows, err := exec.QueryContext(ctx, query, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch returned books: %w", err)
	}
	defer rows.Close()

	var result []models.ReturnedBook
	for rows.Next() {
		var rb models.ReturnedBook
		if err := rows.Scan(&rb.ID, &rb.BookID, &rb.LibraryID, &rb.MemberID, &rb.ReturnedAt); err != nil {
			return nil, fmt.Errorf("failed to scan returned book row: %w", err)
		}
		result = append(result, rb)
	}
	return result, rows.Err()
}

// GetReturnedBook fetches a single pending returned-book record scoped to a library+book,
// used to validate the assign_shelf request and to get the member_id for cleanup.
func (r *bookRepository) GetReturnedBook(ctx context.Context, exec GormExecutor, libraryID, bookID int) (*models.ReturnedBook, error) {
	query := `
		SELECT id, book_id, library_id, member_id, returned_at
		FROM returned_books
		WHERE library_id = $1 AND book_id = $2
		ORDER BY returned_at
		LIMIT 1;
	`
	var rb models.ReturnedBook
	err := exec.QueryRowContext(ctx, query, libraryID, bookID).
		Scan(&rb.ID, &rb.BookID, &rb.LibraryID, &rb.MemberID, &rb.ReturnedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch returned book: %w", err)
	}
	return &rb, nil
}

func (r *bookRepository) DeleteReturnedBook(ctx context.Context, exec GormExecutor, id int) error {
	query := `DELETE FROM returned_books WHERE id = $1;`
	_, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete returned book record: %w", err)
	}
	return nil
}

func (r *bookRepository) GetActiveLoan(ctx context.Context, exec GormExecutor, bookID, memberID int) (*models.Loan, error) {
	query := `
        SELECT id, book_id, member_id, borrowed_library_id, borrowed_at, borrowed_days
        FROM loans
        WHERE book_id = $1 AND member_id = $2 AND returned_at IS NULL`

	loan := &models.Loan{}
	err := exec.QueryRowContext(ctx, query, bookID, memberID).Scan(
		&loan.ID, &loan.BookID, &loan.MemberID, &loan.BorrowedLibraryID, &loan.BorrowedAt, &loan.BorrowedDays,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch active loan: %w", err)
	}
	return loan, nil
}
