package services

import (
	"context"
	"database/sql"
	"errors"
	"library/internal/dto"
	"library/internal/models"
	"library/internal/repositories"
)

type MemberService interface {
	RegisterMember(ctx context.Context, req dto.CreateMemberRequest) (*models.BookMember, error)
	GetMemberByID(ctx context.Context, id int64) (*models.BookMember, error)
	ListMembers(ctx context.Context) ([]models.BookMember, error)
	UpdateMember(ctx context.Context, id int64, req dto.CreateMemberRequest) (*models.BookMember, error)
	DeleteMember(ctx context.Context, id int64) error
}

type memberService struct {
	db   *sql.DB
	repo repositories.MemberRepository
}

func NewMemberService(db *sql.DB, repo repositories.MemberRepository) MemberService {
	return &memberService{db: db, repo: repo}
}

func (s *memberService) RegisterMember(ctx context.Context, req dto.CreateMemberRequest) (*models.BookMember, error) {
	existing, err := s.repo.GetByEmail(ctx, s.db, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("a library member profile with this email address already exists")
	}

	member := &models.BookMember{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.repo.Create(ctx, s.db, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *memberService) GetMemberByID(ctx context.Context, id int64) (*models.BookMember, error) {
	member, err := s.repo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member record not found")
	}
	return member, nil
}

func (s *memberService) ListMembers(ctx context.Context) ([]models.BookMember, error) {
	return s.repo.List(ctx, s.db)
}

func (s *memberService) UpdateMember(ctx context.Context, id int64, req dto.CreateMemberRequest) (*models.BookMember, error) {
	member, err := s.repo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("target member profile not found")
	}

	if req.Email != member.Email {
		duplicate, err := s.repo.GetByEmail(ctx, s.db, req.Email)
		if err != nil {
			return nil, err
		}
		if duplicate != nil {
			return nil, errors.New("this updated email destination is already claimed by another user")
		}
	}

	member.Name = req.Name
	member.Email = req.Email

	if err := s.repo.Update(ctx, s.db, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *memberService) DeleteMember(ctx context.Context, id int64) error {
	hasLoans, err := s.repo.HasOutstandingLoans(ctx, s.db, id)
	if err != nil {
		return err
	}
	if hasLoans {
		return errors.New("cannot delete member profile: account currently has outstanding unreturned book loans")
	}

	return s.repo.Delete(ctx, s.db, id)
}
