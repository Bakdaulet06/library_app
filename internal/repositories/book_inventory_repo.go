package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/models"
)

type BookInventoryRepository interface {
	GetAvailableCopies(ctx context.Context, exec GormExecutor, book_id, library_Id int) (*int, error)
	CreateOrUpdate(ctx context.Context, exec GormExecutor, b *models.BookInventory) error
	List(ctx context.Context, exec GormExecutor) ([]models.BookInventory, error)
	Delete(ctx context.Context, exec GormExecutor, libraryId, bookId int) error

	DecrementInventory(ctx context.Context, exec GormExecutor, bookID, libraryID int) error
	IncrementInventory(ctx context.Context, exec GormExecutor, bookID, libraryID int) error
}

type bookInventoryRepository struct{}

func NewBookInventoryRepository() BookInventoryRepository {
	return &bookInventoryRepository{}
}

func (r *bookInventoryRepository) GetAvailableCopies(ctx context.Context, exec GormExecutor, bookId, libraryId int) (*int, error) {
	// 1. Fixed parameter placeholder from &2 -> $2
	query := `SELECT available_copies FROM book_inventory WHERE book_id = $1 AND library_id = $2`

	var copies int
	err := exec.QueryRowContext(ctx, query, bookId, libraryId).Scan(&copies)

	// 2. Return nil, nil if the book isn't stocked at this library at all
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// 3. Cleanly return pointer to copies count
	return &copies, nil
}

func (r *bookInventoryRepository) CreateOrUpdate(ctx context.Context, exec GormExecutor, b *models.BookInventory) error {
	query := `
		INSERT INTO book_inventory (library_id, book_id, available_copies)
		VALUES ($1, $2, $3)
		ON CONFLICT (library_id, book_id) 
		DO UPDATE SET available_copies = EXCLUDED.available_copies`

	_, err := exec.ExecContext(ctx, query, b.LibraryID, b.BookID, b.AvailableCopies)
	if err != nil {
		return fmt.Errorf("failed to create or update book inventory: %w", err)
	}

	return nil
}

// --- 2. List ---
// Fetches all inventory stock records across all libraries
func (r *bookInventoryRepository) List(ctx context.Context, exec GormExecutor) ([]models.BookInventory, error) {
	query := `
		SELECT library_id, book_id, available_copies 
		FROM book_inventory 
		ORDER BY library_id, book_id`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query book inventory: %w", err)
	}
	defer rows.Close()

	var inventory []models.BookInventory

	for rows.Next() {
		var item models.BookInventory
		if err := rows.Scan(&item.LibraryID, &item.BookID, &item.AvailableCopies); err != nil {
			return nil, fmt.Errorf("failed to scan book inventory row: %w", err)
		}
		inventory = append(inventory, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during book inventory row iteration: %w", err)
	}

	return inventory, nil
}

// --- 3. Delete ---
// Deletes a specific book stock entry from a library branch
func (r *bookInventoryRepository) Delete(ctx context.Context, exec GormExecutor, libraryID, bookID int) error {
	query := `DELETE FROM book_inventory WHERE library_id = $1 AND book_id = $2`

	result, err := exec.ExecContext(ctx, query, libraryID, bookID)
	if err != nil {
		return fmt.Errorf("failed to execute inventory delete query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no inventory record found to delete for library_id %d and book_id %d", libraryID, bookID)
	}

	return nil
}

// Decrements stock count safely (prevents going below 0)
func (r *bookInventoryRepository) DecrementInventory(ctx context.Context, exec GormExecutor, bookID, libraryID int) error {
	query := `
		UPDATE book_inventory 
		SET available_copies = available_copies - 1 
		WHERE book_id = $1 AND library_id = $2 AND available_copies > 0`

	res, err := exec.ExecContext(ctx, query, bookID, libraryID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("failed to decrement stock: copy count was already 0 or inventory record missing")
	}

	return nil
}

// Increments stock count safely using UPSERT (creates stock record if returning to a branch for the first time)
func (r *bookInventoryRepository) IncrementInventory(ctx context.Context, exec GormExecutor, bookID, libraryID int) error {
	query := `
		INSERT INTO book_inventory (library_id, book_id, available_copies)
		VALUES ($1, $2, 1)
		ON CONFLICT (library_id, book_id) 
		DO UPDATE SET available_copies = book_inventory.available_copies + 1`

	_, err := exec.ExecContext(ctx, query, libraryID, bookID)
	return err
}
