package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/dto"
	"library/internal/models"
)

type BookshelfRepository interface {
	Create(ctx context.Context, exec GormExecutor, shelf *models.Bookshelf) error
	GetByID(ctx context.Context, exec GormExecutor, libraryID, shelfID int) (*models.Bookshelf, error)
	GetByLibraryID(ctx context.Context, exec GormExecutor, libraryID int) ([]models.Bookshelf, error)
	GetBooksByShelfID(ctx context.Context, exec GormExecutor, libraryID, shelfID int) ([]dto.BookWithShelfStockResponse, error)
	GetBookByShelfID(ctx context.Context, exec GormExecutor, libraryID, shelfID, bookID int) (*dto.BookWithShelfStockResponse, error)
	UpdateEmptySpace(ctx context.Context, exec GormExecutor, libraryID, shelfID int, spaceDelta int) error
	Delete(ctx context.Context, exec GormExecutor, libraryID, shelfID int) error
}

type bookshelfRepository struct{}

func NewBookshelfRepository() BookshelfRepository {
	return &bookshelfRepository{}
}

// Create inserts a new bookshelf into a specific library.
// empty_space is automatically initialized to match capacity.
func (r *bookshelfRepository) Create(ctx context.Context, exec GormExecutor, shelf *models.Bookshelf) error {
	query := `
		INSERT INTO bookshelves (library_id, code, capacity, empty_space)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := exec.QueryRowContext(ctx, query, shelf.LibraryID, shelf.Code, shelf.Capacity, shelf.Capacity).
		Scan(&shelf.ID, &shelf.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create bookshelf: %w", err)
	}
	return nil
}

// GetByID fetches a specific bookshelf by its primary key ID.
func (r *bookshelfRepository) GetByID(ctx context.Context, exec GormExecutor, libraryID, shelfID int) (*models.Bookshelf, error) {
	query := `
		SELECT id, library_id, code, capacity, empty_space, created_at
		FROM bookshelves
		WHERE library_id = $1 AND id = $2
	`
	var shelf models.Bookshelf
	err := exec.QueryRowContext(ctx, query, libraryID, shelfID).Scan(
		&shelf.ID,
		&shelf.LibraryID,
		&shelf.Code,
		&shelf.Capacity,
		&shelf.EmptySpace,
		&shelf.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bookshelf not found")
		}
		return nil, fmt.Errorf("failed to get bookshelf: %w", err)
	}

	return &shelf, nil
}

// GetByLibraryID lists all bookshelves available inside a specific library branch.
func (r *bookshelfRepository) GetByLibraryID(ctx context.Context, exec GormExecutor, libraryID int) ([]models.Bookshelf, error) {
	query := `
		SELECT id, library_id, code, capacity, empty_space, created_at
		FROM bookshelves
		WHERE library_id = $1
		ORDER BY code ASC
	`
	rows, err := exec.QueryContext(ctx, query, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookshelves: %w", err)
	}
	defer rows.Close()

	var shelves []models.Bookshelf
	for rows.Next() {
		var s models.Bookshelf
		err := rows.Scan(
			&s.ID,
			&s.LibraryID,
			&s.Code,
			&s.Capacity,
			&s.EmptySpace,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookshelf row: %w", err)
		}
		shelves = append(shelves, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return shelves, nil
}

func (r *bookshelfRepository) GetBooksByShelfID(ctx context.Context, exec GormExecutor, libraryID, shelfID int) ([]dto.BookWithShelfStockResponse, error) {
	query := `
		SELECT 
			b.id AS book_id,
			b.title,
			b.author,
			b.isbn,
			b.genre_id,
			bi.available_copies
		FROM book_inventory bi
		JOIN books b ON bi.book_id = b.id
		WHERE bi.library_id = $1 AND bi.bookshelf_id = $2
	`

	rows, err := exec.QueryContext(ctx, query, libraryID, shelfID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []dto.BookWithShelfStockResponse
	for rows.Next() {
		var item dto.BookWithShelfStockResponse
		if err := rows.Scan(&item.BookID, &item.Title, &item.Author, &item.ISBN, &item.GenreID, &item.AvailableCopies); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}

func (r *bookshelfRepository) GetBookByShelfID(ctx context.Context, exec GormExecutor, libraryID, shelfID, bookID int) (*dto.BookWithShelfStockResponse, error) {
	query := `
		SELECT 
			b.id AS book_id,
			b.title,
			b.author,
			b.isbn,
			b.genre_id,
			bi.available_copies
		FROM book_inventory bi
		JOIN books b ON bi.book_id = b.id
		WHERE bi.library_id = $1 AND bi.bookshelf_id = $2 AND bi.book_id = $3
	`

	var resp dto.BookWithShelfStockResponse
	err := exec.QueryRowContext(ctx, query, libraryID, shelfID, bookID).Scan(
		&resp.BookID,
		&resp.Title,
		&resp.Author,
		&resp.ISBN,
		&resp.GenreID,
		&resp.AvailableCopies,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("book with id %d not found on bookshelf %d", bookID, shelfID)
		}
		return nil, fmt.Errorf("failed to fetch book on shelf: %w", err)
	}

	return &resp, nil
}

// UpdateEmptySpace increments or decrements the available shelf space safely.
// Pass negative spaceDelta (e.g. -10) when adding books.
// Pass positive spaceDelta (e.g. +1) when borrowing books (frees up space).
func (r *bookshelfRepository) UpdateEmptySpace(ctx context.Context, exec GormExecutor, libraryID, shelfID int, spaceDelta int) error {
	query := `
		UPDATE bookshelves
		SET empty_space = empty_space - $3
		WHERE id = $2 AND library_id = $1 AND (empty_space - $3 >= 0) AND (empty_space - $3 <= capacity)
	`
	res, err := exec.ExecContext(ctx, query, libraryID, shelfID, spaceDelta)
	if err != nil {
		return fmt.Errorf("failed to update empty space: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("failed to update shelf space: operation exceeds shelf bounds or shelf not found %d", spaceDelta)
	}

	return nil
}

// Delete removes a bookshelf by ID.
func (r *bookshelfRepository) Delete(ctx context.Context, exec GormExecutor, libraryID, shelfID int) error {
	// 1. Delete and return affected rows matching BOTH shelfID AND libraryID
	query := `
		DELETE FROM bookshelves 
		WHERE id = $1 AND library_id = $2 AND empty_space = capacity 
	`
	res, err := exec.ExecContext(ctx, query, shelfID, libraryID)
	if err != nil {
		return fmt.Errorf("failed to delete bookshelf: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// Distinguish between 'not found in this library' vs 'not empty'
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM bookshelves WHERE id = $1 AND library_id = $2)`

		if err := exec.QueryRowContext(ctx, checkQuery, shelfID, libraryID).Scan(&exists); err == nil && exists {
			return fmt.Errorf("cannot delete bookshelf: it still contains books")
		}

		return fmt.Errorf("bookshelf not found in the specified library")
	}

	return nil
}
