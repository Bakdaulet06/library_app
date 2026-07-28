package repositories

import (
	"context"
	"database/sql"
	"errors"
	"library/internal/models"
)

type EmployeeRepository interface {
	Create(ctx context.Context, exec GormExecutor, emp *models.Employee) error
	GetByMemberID(ctx context.Context, exec GormExecutor, memberID int) (*models.Employee, error)
	List(ctx context.Context, exec GormExecutor) ([]models.Employee, error)
	Update(ctx context.Context, exec GormExecutor, emp *models.Employee) error
	Delete(ctx context.Context, exec GormExecutor, memberID int) error
}

type employeeRepository struct{}

func NewEmployeeRepository() EmployeeRepository {
	return &employeeRepository{}
}

func (r *employeeRepository) Create(ctx context.Context, exec GormExecutor, emp *models.Employee) error {
	query := `
		INSERT INTO employees (member_id, position, salary, library_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return exec.QueryRowContext(ctx, query, emp.MemberID, emp.Position, emp.Salary, emp.LibraryID).Scan(&emp.ID)
}

func (r *employeeRepository) GetByMemberID(ctx context.Context, exec GormExecutor, memberID int) (*models.Employee, error) {
	query := `
		SELECT 
			e.id, e.member_id, e.position, e.salary, e.library_id,
			m.email, m.role, m.joined_at
		FROM employees e
		JOIN members m ON e.member_id = m.id
		WHERE e.member_id = $1`

	var emp models.Employee
	err := exec.QueryRowContext(ctx, query, memberID).Scan(
		&emp.ID, &emp.MemberID, &emp.Position, &emp.Salary, &emp.LibraryID,
		&emp.Member.Email, &emp.Member.Role, &emp.Member.JoinedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	emp.Member.ID = memberID
	return &emp, nil
}

func (r *employeeRepository) List(ctx context.Context, exec GormExecutor) ([]models.Employee, error) {
	query := `
		SELECT 
			e.id, e.member_id, e.position, e.salary, e.library_id,
			m.email, m.role, m.joined_at
		FROM employees e
		JOIN members m ON e.member_id = m.id
		ORDER BY e.id ASC`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var emp models.Employee
		err := rows.Scan(
			&emp.ID, &emp.MemberID, &emp.Position, &emp.Salary, &emp.LibraryID,
			&emp.Member.Email, &emp.Member.Role, &emp.Member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		emp.Member.ID = emp.MemberID
		employees = append(employees, emp)
	}

	return employees, nil
}

func (r *employeeRepository) Update(ctx context.Context, exec GormExecutor, emp *models.Employee) error {
	query := `UPDATE employees SET position = $1, salary = $2 WHERE member_id = $3`
	res, err := exec.ExecContext(ctx, query, emp.Position, emp.Salary, emp.MemberID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("employee record not found for update operations")
	}
	return err
}

func (r *employeeRepository) Delete(ctx context.Context, exec GormExecutor, memberID int) error {
	query := `DELETE FROM members WHERE id = $1`
	res, err := exec.ExecContext(ctx, query, memberID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.New("employee profile record not found for deletion operations")
	}
	return err
}
