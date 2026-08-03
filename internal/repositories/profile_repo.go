package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"library/internal/models"
)

type ProfileRepository interface {
	GetUserByID(ctx context.Context, exec GormExecutor, userID int) (*models.BookMember, error)
	UpdateUser(ctx context.Context, exec GormExecutor, userID int, email, hashedPassword string) (*models.BookMember, error)
	HasActiveLoan(ctx context.Context, exec GormExecutor, userID int) (bool, error)
	DeleteUser(ctx context.Context, exec GormExecutor, userID int) error
	DeleteCard(ctx context.Context, exec GormExecutor, userID int) error

	GetCardByUserID(ctx context.Context, exec GormExecutor, userID int) (*models.Card, error)
	CreateCard(ctx context.Context, exec GormExecutor, userID int) (*models.Card, error)
	UpdateCardAmountWithTx(ctx context.Context, tx *sql.Tx, userID int, amountChange float64) (*models.Card, error)
	UpdateCardBalance(ctx context.Context, exec GormExecutor, cardID int, delta float64) error
}

type profileRepository struct{}

func NewProfileRepository() ProfileRepository {
	return &profileRepository{}
}

// GetUserByID retrieves user profile data
func (r *profileRepository) GetUserByID(ctx context.Context, exec GormExecutor, userID int) (*models.BookMember, error) {
	query := `SELECT id, email, password, role, joined_at FROM members WHERE id = $1`

	user := &models.BookMember{}
	err := exec.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.JoinedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return user, nil
}

// UpdateUser uses the UpdateClientProfile DTO to update user email
func (r *profileRepository) UpdateUser(ctx context.Context, exec GormExecutor, userID int, email, hashedPassword string) (*models.BookMember, error) {
	query := `
		UPDATE members 
		SET email = $1, password = $2 
		WHERE id = $3 
		RETURNING id, email, role, joined_at`

	updated := &models.BookMember{}
	err := exec.QueryRowContext(ctx, query, email, hashedPassword, userID).Scan(
		&updated.ID,
		&updated.Email,
		&updated.Role,
		&updated.JoinedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	return updated, nil
}

// DeleteUser removes the user profile
func (r *profileRepository) DeleteUser(ctx context.Context, exec GormExecutor, userID int) error {
	query := `DELETE FROM members WHERE id = $1`
	res, err := exec.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("user not found or already deleted")
	}
	return nil
}

func (r *profileRepository) DeleteCard(ctx context.Context, exec GormExecutor, userID int) error {
	query := `
		DELETE FROM cards
		WHERE member_id = $1 AND amount = 0.00`

	res, err := exec.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete card for user %d: %w", userID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		// Either card doesn't exist OR it has money on it
		return fmt.Errorf("cannot delete card: card not found or still has remaining balance")
	}

	return nil
}

func (r *profileRepository) HasActiveLoan(ctx context.Context, exec GormExecutor, userID int) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM loans 
			WHERE member_id = $1 AND returned_at IS NULL
		)`

	var hasActive bool
	err := exec.QueryRowContext(ctx, query, userID).Scan(&hasActive)
	if err != nil {
		return false, fmt.Errorf("failed to check active loans: %w", err)
	}

	return hasActive, nil
}

// ----------------------------------------------------
// Card Operations
// ----------------------------------------------------

// GetCardByUserID retrieves card details for a given member
func (r *profileRepository) GetCardByUserID(ctx context.Context, exec GormExecutor, userID int) (*models.Card, error) {
	query := `SELECT id, member_id, amount, created_at FROM cards WHERE member_id = $1`

	card := &models.Card{}
	err := exec.QueryRowContext(ctx, query, userID).Scan(
		&card.ID,
		&card.MemberID,
		&card.Amount,
		&card.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found for user")
		}
		return nil, fmt.Errorf("failed to fetch card details: %w", err)
	}
	return card, nil
}

// In ProfileRepository interface:
// CreateCard(ctx context.Context, db *sql.DB, userID int) (*models.Card, error)

func (r *profileRepository) CreateCard(ctx context.Context, exec GormExecutor, userID int) (*models.Card, error) {
	query := `
		INSERT INTO cards (member_id, amount)
		VALUES ($1, 0.00)
		RETURNING id, member_id, amount, created_at`

	card := &models.Card{}
	err := exec.QueryRowContext(ctx, query, userID).Scan(
		&card.ID,
		&card.MemberID,
		&card.Amount,
		&card.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	return card, nil
}

// UpdateCardAmountWithTx executes a card balance update inside an active transaction
// FOR UPDATE locks the row to prevent concurrent balance manipulation race conditions
func (r *profileRepository) UpdateCardAmountWithTx(ctx context.Context, tx *sql.Tx, userID int, amountChange float64) (*models.Card, error) {
	selectQuery := `SELECT id, member_id, amount, created_at FROM cards WHERE member_id = $1 FOR UPDATE`

	card := &models.Card{}
	err := tx.QueryRowContext(ctx, selectQuery, userID).Scan(
		&card.ID,
		&card.MemberID,
		&card.Amount,
		&card.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found")
		}
		return nil, err
	}

	newBalance := card.Amount + amountChange
	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient card funds (current balance: $%.2f)", card.Amount)
	}

	updateQuery := `UPDATE cards SET amount = $1 WHERE member_id = $2 RETURNING id, member_id, amount, created_at`
	updatedCard := &models.Card{}
	err = tx.QueryRowContext(ctx, updateQuery, newBalance, userID).Scan(
		&updatedCard.ID,
		&updatedCard.MemberID,
		&updatedCard.Amount,
		&updatedCard.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update card amount: %w", err)
	}

	return updatedCard, nil
}

// UpdateCardBalance adjusts a card's balance by delta (can be negative).
// Balance is allowed to go below zero.
func (r *profileRepository) UpdateCardBalance(ctx context.Context, exec GormExecutor, cardID int, delta float64) error {
	query := `UPDATE cards SET amount = amount + $1 WHERE id = $2`
	_, err := exec.ExecContext(ctx, query, delta, cardID)
	if err != nil {
		return fmt.Errorf("failed to update card balance: %w", err)
	}
	return nil
}
