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

type EmployeeService interface {
	RegisterEmployee(ctx context.Context, req dto.CreateEmployeeRequest) (*models.Employee, error)
	GetEmployeeByMemberID(ctx context.Context, memberID int) (*models.Employee, error)
	ListEmployees(ctx context.Context) ([]models.Employee, error)
	UpdateEmployee(ctx context.Context, memberID int, req dto.UpdateEmployeeRequest) (*models.Employee, error)
	DeleteEmployee(ctx context.Context, memberID int) error
}

type employeeService struct {
	db           *sql.DB
	employeeRepo repositories.EmployeeRepository
	memberRepo   repositories.MemberRepository
	userRepo     repositories.UserRepository
	libraryRepo  repositories.LibraryRepository
}

func NewEmployeeService(db *sql.DB, employeeRepo repositories.EmployeeRepository, memberRepo repositories.MemberRepository, userRepo repositories.UserRepository, libraryRepo repositories.LibraryRepository) EmployeeService {
	return &employeeService{
		db:           db,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		memberRepo:   memberRepo,
		libraryRepo:  libraryRepo,
	}
}

func (s *employeeService) RegisterEmployee(ctx context.Context, req dto.CreateEmployeeRequest) (*models.Employee, error) {
	existing, err := s.memberRepo.GetByEmail(ctx, s.db, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("a library member profile with this email address already exists")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// pre-check inside the transaction, so it sees a consistent view
	hasEmployee, err := s.libraryRepo.HasEmployee(ctx, tx, req.LibraryID)
	if err != nil {
		return nil, err
	}
	if hasEmployee {
		return nil, errors.New("this library already has an assigned employee")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	member := &models.BookMember{
		Email:    req.Email,
		Password: string(hashedBytes),
		Role:     "employee",
	}
	if err := s.userRepo.Create(ctx, tx, member); err != nil {
		return nil, err
	}

	emp := &models.Employee{
		MemberID:  member.ID,
		Position:  req.Position,
		Salary:    req.Salary,
		LibraryID: req.LibraryID,
	}
	if err := s.employeeRepo.Create(ctx, tx, emp); err != nil {
		return nil, err
	}

	if err := s.libraryRepo.CreateLibraryEmployee(ctx, tx, req.LibraryID, member.ID); err != nil {
		if errors.Is(err, repositories.ErrLibraryAlreadyHasEmployee) {
			return nil, errors.New("this library already has an assigned employee")
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	emp.Member = *member
	return emp, nil
}

func (s *employeeService) GetEmployeeByMemberID(ctx context.Context, memberID int) (*models.Employee, error) {
	emp, err := s.employeeRepo.GetByMemberID(ctx, s.db, memberID)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, errors.New("employee profile record not found")
	}
	return emp, nil
}

func (s *employeeService) ListEmployees(ctx context.Context) ([]models.Employee, error) {
	return s.employeeRepo.List(ctx, s.db)
}

func (s *employeeService) UpdateEmployee(ctx context.Context, memberID int, req dto.UpdateEmployeeRequest) (*models.Employee, error) {
	emp, err := s.employeeRepo.GetByMemberID(ctx, s.db, memberID)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, errors.New("target employee profile not found")
	}

	emp.Position = req.Position
	emp.Salary = req.Salary

	if err := s.employeeRepo.Update(ctx, s.db, emp); err != nil {
		return nil, err
	}
	return emp, nil
}

func (s *employeeService) DeleteEmployee(ctx context.Context, memberID int) error {
	return s.employeeRepo.Delete(ctx, s.db, memberID)
}
