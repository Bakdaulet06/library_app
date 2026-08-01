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

type MemberService interface {
	GetMemberByID(ctx context.Context, id int) (*models.BookMember, error)
	ListMembers(ctx context.Context) ([]models.BookMember, error)
	UpdateMember(ctx context.Context, id int, req dto.RegisterRequest) (*models.BookMember, error)
	DeleteMember(ctx context.Context, id int) error
	GetMemberLoans(ctx context.Context, memberID int) ([]models.Loan, error)
}

type memberService struct {
	db         *sql.DB
	memberRepo repositories.MemberRepository
	bookRepo   repositories.BookRepository
}

func NewMemberService(db *sql.DB, memberRepo repositories.MemberRepository, bookRepo repositories.BookRepository) MemberService {
	return &memberService{db: db, memberRepo: memberRepo, bookRepo: bookRepo}
}

func (s *memberService) GetMemberByID(ctx context.Context, id int) (*models.BookMember, error) {
	member, err := s.memberRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member record not found")
	}
	return member, nil
}

func (s *memberService) ListMembers(ctx context.Context) ([]models.BookMember, error) {
	return s.memberRepo.List(ctx, s.db)
}

func (s *memberService) UpdateMember(ctx context.Context, id int, req dto.RegisterRequest) (*models.BookMember, error) {
	member, err := s.memberRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("target member profile not found")
	}

	if req.Email != member.Email {
		duplicate, err := s.memberRepo.GetByEmail(ctx, s.db, req.Email)
		if err != nil {
			return nil, err
		}
		if duplicate != nil {
			return nil, errors.New("this updated email destination is already claimed by another user")
		}
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	member.Email = req.Email
	member.Password = string(hashedBytes)

	if err := s.memberRepo.Update(ctx, s.db, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *memberService) DeleteMember(ctx context.Context, id int) error {
	hasLoans, err := s.memberRepo.HasOutstandingLoans(ctx, s.db, id)
	if err != nil {
		return err
	}
	if hasLoans {
		return errors.New("cannot delete member profile: account currently has outstanding unreturned book loans")
	}

	return s.memberRepo.Delete(ctx, s.db, id)
}

func (s *memberService) GetMemberLoans(ctx context.Context, memberID int) ([]models.Loan, error) {

	fmt.Println("memberRepo:", s.memberRepo)
	fmt.Println("bookRepo:", s.bookRepo)
	fmt.Println("db:", s.db)
	// 1. Validate that member exists
	if _, err := s.memberRepo.GetByID(ctx, s.db, memberID); err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// 2. Query loans via bookRepo
	loans, err := s.bookRepo.GetLoansByMemberID(ctx, s.db, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch member loans: %w", err)
	}

	return loans, nil
}
