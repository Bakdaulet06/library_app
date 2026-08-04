package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/params"
	"library/internal/repositories"
)

type GenreService interface {
	CreateGenre(ctx context.Context, req dto.CreateOrUpdateGenreRequest) (*models.Genre, error)
	GetGenreByID(ctx context.Context, id int) (*models.Genre, error)
	GetAllGenres(ctx context.Context, p params.Pagination) ([]models.Genre, error)
	UpdateGenre(ctx context.Context, id int, req dto.CreateOrUpdateGenreRequest) (*models.Genre, error)
	DeleteGenre(ctx context.Context, id int) error
}

type genreService struct {
	db   *sql.DB
	repo repositories.GenreRepository
}

func NewGenreService(db *sql.DB, repo repositories.GenreRepository) GenreService {
	return &genreService{db: db, repo: repo}
}

func (s *genreService) CreateGenre(ctx context.Context, req dto.CreateOrUpdateGenreRequest) (*models.Genre, error) {
	exists, err := s.repo.ExistsByName(ctx, s.db, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check genre: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("Genre with this name already exists")
	}
	genre := &models.Genre{
		Name: req.Name,
	}
	if err := s.repo.Create(ctx, s.db, genre); err != nil {
		return nil, err
	}
	return genre, nil
}

func (s *genreService) GetGenreByID(ctx context.Context, id int) (*models.Genre, error) {
	exists, err := s.repo.ExistsByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check genre: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("Genre with this id doesn't exist")
	}
	return s.repo.GetByID(ctx, s.db, id)
}

func (s *genreService) GetAllGenres(ctx context.Context, p params.Pagination) ([]models.Genre, error) {
	genres, err := s.repo.GetAll(ctx, s.db, p)
	if err != nil {
		return nil, err
	}

	if genres == nil {
		return []models.Genre{}, nil
	}

	return genres, nil
}

func (s *genreService) UpdateGenre(ctx context.Context, id int, req dto.CreateOrUpdateGenreRequest) (*models.Genre, error) {
	genre := &models.Genre{
		ID:   id,
		Name: req.Name,
	}

	if err := s.repo.Update(ctx, s.db, genre); err != nil {
		return nil, err
	}
	return genre, nil
}

func (s *genreService) DeleteGenre(ctx context.Context, id int) error {
	// 1. Check if any books are associated with this genre
	hasBookWithThisGenre, err := s.repo.HasBookWithThisGenre(ctx, s.db, id)
	if err != nil {
		return fmt.Errorf("failed to check genre usage: %w", err)
	}

	// 2. Prevent deletion if books are using this genre
	if hasBookWithThisGenre {
		return errors.New("cannot delete genre: it is assigned to one or more books") // Return domain error for HTTP 400/409 handler
	}

	// 3. Delete the genre
	if err := s.repo.Delete(ctx, s.db, id); err != nil {
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	return nil
}
