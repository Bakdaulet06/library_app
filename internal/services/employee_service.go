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

	canAdd, err := s.libraryRepo.CanAddEmployee(ctx, tx, req.LibraryID, req.Position)
	if err != nil {
		return nil, err
	}
	if !canAdd {
		return nil, fmt.Errorf("library %d has already reached maximum capacity for position %s", req.LibraryID, req.Position)
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

	if err := s.libraryRepo.CreateLibraryEmployee(ctx, tx, req.LibraryID, emp.MemberID, emp.Position); err != nil {
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
	// 1. Begin Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 2. Fetch existing employee using transaction
	emp, err := s.employeeRepo.GetByMemberID(ctx, tx, memberID)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, errors.New("target employee profile not found")
	}

	oldPosition := emp.Position
	newPosition := req.Position

	// 3. Only check capacity if the position is actually CHANGING
	if oldPosition != newPosition {
		canAdd, err := s.libraryRepo.CanAddEmployee(ctx, tx, emp.LibraryID, newPosition)
		if err != nil {
			return nil, err
		}
		if !canAdd {
			return nil, fmt.Errorf("library %d has already reached maximum capacity for position %s", emp.LibraryID, newPosition)
		}
	}

	// 4. Update struct fields
	emp.Position = newPosition
	emp.Salary = req.Salary

	// 5. Update main employee table
	if err := s.employeeRepo.Update(ctx, tx, emp); err != nil {
		return nil, err
	}

	// 6. Update junction table if position changed
	if oldPosition != newPosition {
		err := s.libraryRepo.UpdateLibraryEmployeePosition(ctx, tx, emp.LibraryID, emp.MemberID, oldPosition, newPosition)
		if err != nil {
			return nil, fmt.Errorf("failed to update library employee position: %w", err)
		}
	}

	// 7. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *employeeService) DeleteEmployee(ctx context.Context, memberID int) error {
	return s.employeeRepo.Delete(ctx, s.db, memberID)
}
