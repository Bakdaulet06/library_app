package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type ProfileService interface {
	GetProfileByUserID(ctx context.Context, userID int) (*models.BookMember, error)
	UpdateProfile(ctx context.Context, userID int, req dto.UpdateClientProfile) (*models.BookMember, error)
	DeleteProfile(ctx context.Context, userID int) error
	DeleteCard(ctx context.Context, userID int) error
	GetCardByUserID(ctx context.Context, userID int) (*models.Card, error)
	CreateCard(ctx context.Context, userID int) (*models.Card, error)
	DepositMoney(ctx context.Context, userID int, amount float64) (*models.Card, error)
	WithdrawMoney(ctx context.Context, userID int, amount float64) (*models.Card, error)
}

type profileService struct {
	db          *sql.DB
	profileRepo repositories.ProfileRepository
}

func NewProfileService(db *sql.DB, profileRepo repositories.ProfileRepository) ProfileService {
	return &profileService{
		db:          db,
		profileRepo: profileRepo,
	}
}

func (s *profileService) GetProfileByUserID(ctx context.Context, userID int) (*models.BookMember, error) {
	return s.profileRepo.GetUserByID(ctx, s.db, userID)
}

func (s *profileService) UpdateProfile(ctx context.Context, userID int, req dto.UpdateClientProfile) (*models.BookMember, error) {
	// 1. Validate DTO
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 2. Fetch existing user from DB to verify old password
	currentUser, err := s.profileRepo.GetUserByID(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 3. Compare OldPassword with stored bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(req.OldPassword)); err != nil {
		return nil, fmt.Errorf("invalid current password")
	}

	// 4. Hash the NewPassword
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash new password: %w", err)
	}

	// 5. Update email and new hashed password in DB
	return s.profileRepo.UpdateUser(ctx, s.db, userID, req.Email, string(hashedPassword))
}

func (s *profileService) DeleteProfile(ctx context.Context, userID int) error {
	// 1. Check if user has active unreturned loans first
	hasActiveLoan, err := s.profileRepo.HasActiveLoan(ctx, s.db, userID)
	if err != nil {
		return fmt.Errorf("failed to check active loans: %w", err)
	}
	if hasActiveLoan {
		return fmt.Errorf("cannot delete profile: you have active book loans that must be returned first")
	}

	// 2. Check if user has a card and whether it still holds funds
	card, err := s.profileRepo.GetCardByUserID(ctx, s.db, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check user card: %w", err)
	}

	// Block deletion if they still have funds on their card
	if card != nil && card.Amount > 0 {
		return fmt.Errorf("cannot delete profile: withdraw your remaining balance of $%.2f first", card.Amount)
	}

	// 3. Begin transaction for atomic deletion
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 4. Delete the card if one exists
	if card != nil {
		if err := s.profileRepo.DeleteCard(ctx, tx, userID); err != nil {
			return fmt.Errorf("failed to delete user card: %w", err)
		}
	}

	// 5. Delete the user profile
	if err := s.profileRepo.DeleteUser(ctx, tx, userID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	// 6. Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *profileService) DeleteCard(ctx context.Context, userID int) error {
	// 1. Fetch user's card to verify existence and check balance
	card, err := s.profileRepo.GetCardByUserID(ctx, s.db, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("card not found")
		}
		return fmt.Errorf("failed to fetch card: %w", err)
	}

	// 2. Prevent card deletion if funds still remain
	if card.Amount > 0 {
		return fmt.Errorf("cannot delete card: you still have a remaining balance of $%.2f", card.Amount)
	}

	// 3. Delete the card using database transaction for consistency
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.profileRepo.DeleteCard(ctx, tx, userID); err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *profileService) GetCardByUserID(ctx context.Context, userID int) (*models.Card, error) {
	return s.profileRepo.GetCardByUserID(ctx, s.db, userID)
}

func (s *profileService) CreateCard(ctx context.Context, userID int) (*models.Card, error) {
	// Prevent duplicate cards if one already exists
	existingCard, err := s.profileRepo.GetCardByUserID(ctx, s.db, userID)
	if err == nil && existingCard != nil {
		return nil, fmt.Errorf("user already has an active card")
	}

	return s.profileRepo.CreateCard(ctx, s.db, userID)
}

func (s *profileService) DepositMoney(ctx context.Context, userID int, amount float64) (*models.Card, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("deposit amount must be greater than zero")
	}

	// Start transaction using service DB instance
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Pass tx to repository
	card, err := s.profileRepo.UpdateCardAmountWithTx(ctx, tx, userID, amount)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return card, nil
}

func (s *profileService) WithdrawMoney(ctx context.Context, userID int, amount float64) (*models.Card, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("withdrawal amount must be greater than zero")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Pass negative amount for withdrawal
	card, err := s.profileRepo.UpdateCardAmountWithTx(ctx, tx, userID, -amount)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return card, nil
}
