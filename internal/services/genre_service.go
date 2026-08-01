package services

import (
	"context"
	"database/sql"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"
)

type GenreService interface {
	CreateGenre(ctx context.Context, req dto.CreateOrUpdateGenreRequest) (*models.Genre, error)
	GetGenreByID(ctx context.Context, id int) (*models.Genre, error)
	GetAllGenres(ctx context.Context) ([]models.Genre, error)
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
	genre := &models.Genre{
		Name: req.Name,
	}
	if err := s.repo.Create(ctx, s.db, genre); err != nil {
		return nil, err
	}
	return genre, nil
}

func (s *genreService) GetGenreByID(ctx context.Context, id int) (*models.Genre, error) {
	return s.repo.GetByID(ctx, s.db, id)
}

func (s *genreService) GetAllGenres(ctx context.Context) ([]models.Genre, error) {
	return s.repo.GetAll(ctx, s.db)
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
	return s.repo.Delete(ctx, s.db, id)
}
