package repositories

import (
	"context"
	"database/sql"
	"errors"
	"library/internal/models"
)

type MemberRepository interface {
	Create(ctx context.Context, db DBTX, member *models.Member) error
	GetByID(ctx context.Context, db DBTX, id int) (*models.Member, error)
	GetByEmail(ctx context.Context, db DBTX, email string) (*models.Member, error)
}

type memberRepository struct{}

func NewMemberRepository() MemberRepository {
	return &memberRepository{}
}

func (r *memberRepository) Create(ctx context.Context, db DBTX, m *models.Member) error {
	query := `
		INSERT INTO members (name, email)
		VALUES ($1, $2)
		RETURNING id, joined_at;`
	err := db.QueryRowContext(ctx, query, m.Name, m.Email).Scan(&m.ID, &m.JoinedAt)
	return err
}

func (r *memberRepository) GetByID(ctx context.Context, db DBTX, id int) (*models.Member, error) {
	query := `SELECT id, name, email, joined_at FROM members WHERE id = $1;`
	var m models.Member
	err := db.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Email, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *memberRepository) GetByEmail(ctx context.Context, db DBTX, email string) (*models.Member, error) {
	query := `SELECT id, name, email, joined_at FROM members WHERE email = $1;`
	var m models.Member
	err := db.QueryRowContext(ctx, query, email).Scan(&m.ID, &m.Name, &m.Email, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}
