package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/auth"
	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*models.BookMember, error)
	Login(ctx context.Context, req dto.RegisterClientAndLoginRequest) (*dto.AuthResponse, error)
	GetUserByID(ctx context.Context, userID int) (*models.BookMember, error)
	RegisterClient(ctx context.Context, req dto.RegisterClientAndLoginRequest) (*models.BookMember, error)
}

type userService struct {
	db       *sql.DB
	userRepo repositories.UserRepository
}

func NewUserService(db *sql.DB, userRepo repositories.UserRepository) UserService {
	return &userService{db: db, userRepo: userRepo}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (*models.BookMember, error) {
	// 1. Default role to "user" if not specified
	if req.Role == "" {
		req.Role = "client"
	}

	// 2. Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.BookMember{
		Email:    req.Email,
		Password: string(hashedBytes),
		Role:     req.Role,
	}

	// 3. Save to database via repository
	err = s.userRepo.Create(ctx, s.db, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) RegisterClient(ctx context.Context, req dto.RegisterClientAndLoginRequest) (*models.BookMember, error) {
	// 2. Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.BookMember{
		Email:    req.Email,
		Password: string(hashedBytes),
		Role:     "client",
	}

	// 3. Save to database via repository
	err = s.userRepo.Create(ctx, s.db, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, req dto.RegisterClientAndLoginRequest) (*dto.AuthResponse, error) {
	// 1. Fetch user by email
	user, err := s.userRepo.GetByEmail(ctx, s.db, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Compare hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate JWT Token using UserID and Role
	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID int) (*models.BookMember, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// 1. Fetch user from repository
	user, err := s.userRepo.GetByID(ctx, s.db, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}
