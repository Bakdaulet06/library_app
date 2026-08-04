package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"library/internal/models"
	"library/internal/params"
)

var (
	ErrGenreNotFound = errors.New("genre not found")
	ErrGenreExists   = errors.New("genre name already exists")
)

type GenreRepository interface {
	Create(ctx context.Context, exec GormExecutor, genre *models.Genre) error
	GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Genre, error)
	GetAll(ctx context.Context, exec GormExecutor, p params.Pagination) ([]models.Genre, error)
	Update(ctx context.Context, exec GormExecutor, genre *models.Genre) error
	Delete(ctx context.Context, exec GormExecutor, id int) error

	ExistsByID(ctx context.Context, exec GormExecutor, id int) (bool, error)
	ExistsByName(ctx context.Context, exec GormExecutor, name string) (bool, error)
	HasBookWithThisGenre(ctx context.Context, exec GormExecutor, genreID int) (bool, error)
}

type genreRepository struct {
}

func NewGenreRepository() GenreRepository {
	return &genreRepository{}
}

func (r *genreRepository) Create(ctx context.Context, exec GormExecutor, genre *models.Genre) error {
	query := `INSERT INTO genres (name) VALUES ($1) RETURNING id`
	err := exec.QueryRowContext(ctx, query, genre.Name).Scan(&genre.ID)
	if err != nil {
		// Handles unique constraint violations on genre name
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate key") {
			return ErrGenreExists
		}
		return err
	}
	return nil
}

func (r *genreRepository) GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Genre, error) {
	query := `SELECT id, name FROM genres WHERE id = $1`
	var g models.Genre
	err := exec.QueryRowContext(ctx, query, id).Scan(&g.ID, &g.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGenreNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *genreRepository) GetAll(ctx context.Context, exec GormExecutor, p params.Pagination) ([]models.Genre, error) {
	query := `SELECT id, name FROM genres`

	var whereClauses []string
	var args []interface{}
	paramIdx := 1

	// 1. Search by Name (Case-insensitive ILIKE)
	if p.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", paramIdx))
		args = append(args, "%"+p.Search+"%")
		paramIdx++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 2. Allowlist Sorting Columns (Prevents SQL Injection)
	allowedColumns := map[string]string{
		"id":   "id",
		"name": "name",
	}

	sortColumn, exists := allowedColumns[strings.ToLower(p.SortBy)]
	if !exists {
		sortColumn = "id" // Default sorting column
	}

	orderDir := strings.ToUpper(p.Order)
	if orderDir != "DESC" {
		orderDir = "ASC" // Default ordering direction
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, orderDir)

	// 3. Limit & Offset
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, p.Limit, p.Offset)

	// 4. Execute Query
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Ensure empty array [] is returned instead of null in JSON
	if genres == nil {
		genres = []models.Genre{}
	}

	return genres, nil
}

func (r *genreRepository) Update(ctx context.Context, exec GormExecutor, genre *models.Genre) error {
	query := `UPDATE genres SET name = $1 WHERE id = $2`
	res, err := exec.ExecContext(ctx, query, genre.Name, genre.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrGenreNotFound
	}
	return nil
}

func (r *genreRepository) Delete(ctx context.Context, exec GormExecutor, id int) error {
	query := `DELETE FROM genres WHERE id = $1`
	res, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrGenreNotFound
	}
	return nil
}

func (r *genreRepository) HasBookWithThisGenre(ctx context.Context, exec GormExecutor, genreID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM books WHERE genre_id = $1)`

	var exists bool
	err := exec.QueryRowContext(ctx, query, genreID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if books are linked to genre %d: %w", genreID, err)
	}

	return exists, nil
}

func (r *genreRepository) ExistsByID(ctx context.Context, exec GormExecutor, id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM genres WHERE id = $1)`

	var exists bool
	err := exec.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check genre existence by ID %d: %w", id, err)
	}

	return exists, nil
}

// ExistsByName checks if a genre exists by its name (case-insensitive)
func (r *genreRepository) ExistsByName(ctx context.Context, exec GormExecutor, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM genres WHERE LOWER(name) = LOWER($1))`

	var exists bool
	err := exec.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check genre existence by name %s: %w", name, err)
	}

	return exists, nil
}
