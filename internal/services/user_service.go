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
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
}

type userService struct {
	db       *sql.DB
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	// 1. Default role to "user" if not specified
	if req.Role == "" {
		req.Role = "user"
	}

	// 2. Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
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

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
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
