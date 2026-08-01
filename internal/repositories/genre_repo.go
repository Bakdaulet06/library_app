package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"library/internal/models"
)

var (
	ErrGenreNotFound = errors.New("genre not found")
	ErrGenreExists   = errors.New("genre name already exists")
)

type GenreRepository interface {
	Create(ctx context.Context, exec GormExecutor, genre *models.Genre) error
	GetByID(ctx context.Context, exec GormExecutor, id int) (*models.Genre, error)
	GetAll(ctx context.Context, exec GormExecutor) ([]models.Genre, error)
	Update(ctx context.Context, exec GormExecutor, genre *models.Genre) error
	Delete(ctx context.Context, exec GormExecutor, id int) error
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

func (r *genreRepository) GetAll(ctx context.Context, exec GormExecutor) ([]models.Genre, error) {
	query := `SELECT id, name FROM genres ORDER BY id ASC`
	rows, err := exec.QueryContext(ctx, query)
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

	// Always return an empty slice instead of null in JSON
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
