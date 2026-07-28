package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/models"
)

type LibraryRepository interface {
	Create(ctx context.Context, exec GormExecutor, m *models.Library) error
	ListAll(ctx context.Context, exec GormExecutor) ([]models.Library, error)
	ListAllBooks(ctx context.Context, exec GormExecutor, id int) ([]models.LibraryBook, error)
	ListAllLoans(ctx context.Context, exec GormExecutor, id int) ([]models.Loan, error)
	Delete(ctx context.Context, exec GormExecutor, id int) error
	Update(ctx context.Context, exec GormExecutor, m *models.Library) error
	GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Library, error)

	CreateLibraryEmployee(ctx context.Context, exec GormExecutor, libraryID, memberID int) error
	HasEmployee(ctx context.Context, exec GormExecutor, libraryID int) (bool, error)
}

type libraryRepository struct{}

func NewLibraryRepository() LibraryRepository {
	return &libraryRepository{}
}

func (r *libraryRepository) Create(ctx context.Context, exec GormExecutor, m *models.Library) error {
	query := `INSERT INTO libraries (name, address) VALUES ($1, $2) RETURNING id, created_at`
	return exec.QueryRowContext(ctx, query, m.Name, m.Address).Scan(&m.ID, &m.CreatedAt)
}

func (r *libraryRepository) GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Library, error) {
	query := `SELECT id, name, address, created_at FROM libraries WHERE id = $1`
	var m models.Library
	err := exec.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Address, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *libraryRepository) Update(ctx context.Context, exec GormExecutor, m *models.Library) error {
	query := `UPDATE libraries SET name = $1, address = $2 WHERE id = $3`
	res, err := exec.ExecContext(ctx, query, m.Name, m.Address, m.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("library record not found for update operations")
	}
	return err
}

func (r *libraryRepository) Delete(ctx context.Context, exec GormExecutor, id int) error {
	query := `DELETE FROM libraries WHERE id = $1`
	res, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("library record not found for deletion operations")
	}
	return err
}

func (r *libraryRepository) ListAll(ctx context.Context, exec GormExecutor) ([]models.Library, error) {
	query := `SELECT id, name, address, created_at FROM libraries ORDER BY id ASC`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libraries []models.Library
	for rows.Next() {
		var m models.Library
		if err := rows.Scan(&m.ID, &m.Name, &m.Address, &m.CreatedAt); err != nil {
			return nil, err
		}
		libraries = append(libraries, m)
	}
	return libraries, nil
}

func (r *libraryRepository) ListAllBooks(ctx context.Context, exec GormExecutor, libraryID int) ([]models.LibraryBook, error) {
	query := `
		SELECT 
			b.id, b.title, b.author, b.isbn, b.genre_id, b.created_at,
			SUM(bi.available_copies) AS total_available_copies
		FROM books b
		INNER JOIN book_inventory bi ON b.id = bi.book_id
		WHERE bi.library_id = $1
		GROUP BY b.id, b.title, b.author, b.isbn, b.genre_id, b.created_at
		ORDER BY b.id`

	rows, err := exec.QueryContext(ctx, query, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query books for library %d: %w", libraryID, err)
	}
	defer rows.Close()

	var libraryBooks []models.LibraryBook

	for rows.Next() {
		var b models.LibraryBook
		if err := rows.Scan(&b.Book.ID, &b.Book.Title, &b.Book.Author, &b.Book.Isbn, &b.Book.GenreID, &b.Book.CreatedAt, &b.TotalAvailableCopies); err != nil {
			return nil, fmt.Errorf("failed to scan book row: %w", err)
		}
		libraryBooks = append(libraryBooks, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during book rows iteration: %w", err)
	}

	if libraryBooks == nil {
		return []models.LibraryBook{}, nil // Return empty slice instead of nil
	}

	return libraryBooks, nil
}

// --- 2. ListAllLoans ---
// Fetches all loans originated at (borrowed from) a specific library branch ID
func (r *libraryRepository) ListAllLoans(ctx context.Context, exec GormExecutor, libraryID int) ([]models.Loan, error) {
	query := `
		SELECT id, book_id, member_id, borrowed_library_id, returned_library_id, borrowed_at, returned_at
		FROM loans
		WHERE borrowed_library_id = $1
		ORDER BY borrowed_at DESC`

	rows, err := exec.QueryContext(ctx, query, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query loans for library %d: %w", libraryID, err)
	}
	defer rows.Close()

	var loans []models.Loan

	for rows.Next() {
		var l models.Loan
		if err := rows.Scan(
			&l.ID,
			&l.BookID,
			&l.MemberID,
			&l.BorrowedLibraryID,
			&l.ReturnedLibraryID,
			&l.BorrowedAt,
			&l.ReturnedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan loan row: %w", err)
		}
		loans = append(loans, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during loan rows iteration: %w", err)
	}

	if loans == nil {
		return []models.Loan{}, nil // Return empty slice instead of nil
	}

	return loans, nil
}

// HasEmployee returns true if the given library already has an employee assigned.
func (r *libraryRepository) HasEmployee(ctx context.Context, exec GormExecutor, libraryID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM library_employees WHERE library_id = $1)`

	var exists bool
	if err := exec.QueryRowContext(ctx, query, libraryID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check library employee existence: %w", err)
	}
	return exists, nil
}

// CreateLibraryEmployee links a member to a library as its employee.
func (r *libraryRepository) CreateLibraryEmployee(ctx context.Context, exec GormExecutor, libraryID, memberID int) error {
	query := `
		INSERT INTO library_employees (library_id, member_id)
		VALUES ($1, $2);
	`

	_, err := exec.ExecContext(ctx, query, libraryID, memberID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrLibraryAlreadyHasEmployee // define this alongside ErrDuplicateEmail
		}
		return fmt.Errorf("failed to create library employee link: %w", err)
	}
	return nil
}
