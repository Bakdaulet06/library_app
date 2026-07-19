package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"
)

type MemberService interface {
	RegisterMember(ctx context.Context, req dto.CreateMemberRequest) (*models.Member, error)
}

type memberService struct {
	db         *sql.DB
	memberRepo repositories.MemberRepository
}

func NewMemberService(db *sql.DB, mr repositories.MemberRepository) MemberService {
	return &memberService{
		db:         db,
		memberRepo: mr,
	}
}

func (s *memberService) RegisterMember(ctx context.Context, req dto.CreateMemberRequest) (*models.Member, error) {
	// Business Rule: Ensure email uniqueness across account ecosystem
	existing, err := s.memberRepo.GetByEmail(ctx, s.db, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error verifying user profile email uniqueness: %w", err)
	}
	if existing != nil {
		return nil, errors.New("a library user profile is already registered using this email address")
	}

	member := &models.Member{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.memberRepo.Create(ctx, s.db, member); err != nil {
		return nil, fmt.Errorf("failed to persist new member account profile: %w", err)
	}

	return member, nil
}
