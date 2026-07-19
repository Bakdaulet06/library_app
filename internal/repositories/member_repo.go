package repositories

import (
	"context"
	"database/sql"
	"errors"
	"library/internal/models"
)

type MemberRepository interface {
	Create(ctx context.Context, exec GormExecutor, m *models.BookMember) error
	GetByID(ctx context.Context, exec GormExecutor, id int64) (*models.BookMember, error)
	GetByEmail(ctx context.Context, exec GormExecutor, email string) (*models.BookMember, error)
	List(ctx context.Context, exec GormExecutor) ([]models.BookMember, error)
	Update(ctx context.Context, exec GormExecutor, m *models.BookMember) error
	Delete(ctx context.Context, exec GormExecutor, id int64) error
	HasOutstandingLoans(ctx context.Context, exec GormExecutor, id int64) (bool, error)
}

type memberRepository struct{}

func NewMemberRepository() MemberRepository {
	return &memberRepository{}
}

func (r *memberRepository) Create(ctx context.Context, exec GormExecutor, m *models.BookMember) error {
	query := `INSERT INTO members (name, email) VALUES ($1, $2) RETURNING id, joined_at`
	return exec.QueryRowContext(ctx, query, m.Name, m.Email).Scan(&m.ID, &m.JoinedAt)
}

func (r *memberRepository) GetByID(ctx context.Context, exec GormExecutor, id int64) (*models.BookMember, error) {
	query := `SELECT id, name, email, joined_at FROM members WHERE id = $1`
	var m models.BookMember
	err := exec.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Email, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *memberRepository) GetByEmail(ctx context.Context, exec GormExecutor, email string) (*models.BookMember, error) {
	query := `SELECT id, name, email, joined_at FROM members WHERE email = $1`
	var m models.BookMember
	err := exec.QueryRowContext(ctx, query, email).Scan(&m.ID, &m.Name, &m.Email, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *memberRepository) List(ctx context.Context, exec GormExecutor) ([]models.BookMember, error) {
	query := `SELECT id, name, email, joined_at FROM members ORDER BY id ASC`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.BookMember
	for rows.Next() {
		var m models.BookMember
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *memberRepository) Update(ctx context.Context, exec GormExecutor, m *models.BookMember) error {
	query := `UPDATE members SET name = $1, email = $2 WHERE id = $3`
	res, err := exec.ExecContext(ctx, query, m.Name, m.Email, m.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("member profile record not found for update operations")
	}
	return err
}

func (r *memberRepository) Delete(ctx context.Context, exec GormExecutor, id int64) error {
	query := `DELETE FROM members WHERE id = $1`
	res, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("member profile record not found for deletion operations")
	}
	return err
}

func (r *memberRepository) HasOutstandingLoans(ctx context.Context, exec GormExecutor, id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM loans WHERE member_id = $1 AND returned_at IS NULL)`
	var exists bool
	err := exec.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}
