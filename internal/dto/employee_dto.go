package dto

import (
	"errors"
	"fmt"
	"library/internal/models"
	"net/mail"
	"strings"
)

type CreateEmployeeRequest struct {
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	LibraryID int             `json:"library_id"`
	Position  models.Position `json:"position"`
	Salary    float64         `json:"salary"`
}

func (r *CreateEmployeeRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)

	if r.Email == "" {
		return errors.New("email address is required")
	}

	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email address format provided")
	}

	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if r.LibraryID <= 0 {
		return errors.New("invalid library ID")
	}

	if !r.Position.IsValid() {
		return errors.New("invalid or missing employee position")
	}

	// Reject negative numbers explicitly
	if r.Salary < 0 {
		return errors.New("salary cannot be negative")
	}

	// If a custom salary was provided (> 0), make sure it fits within bounds
	if r.Salary > 0 && (r.Salary < models.EmpSalaryMin || r.Salary > models.EmpSalaryMax) {
		return fmt.Errorf("custom salary must be between %.0f and %.0f", models.EmpSalaryMin, models.EmpSalaryMax)
	}

	return nil
}

// SetDefaults assigns baseline values if optional fields are omitted
func (r *CreateEmployeeRequest) SetDefaults() {
	if r.Salary == 0 {
		r.Salary = r.Position.DefaultSalary()
	}
}

type UpdateEmployeeRequest struct {
	Position models.Position `json:"position"` // Updated type to models.Position
	Salary   float64         `json:"salary"`
}

func (r *UpdateEmployeeRequest) Validate() error {
	if !r.Position.IsValid() {
		return errors.New("invalid or missing employee position")
	}

	// Corrected salary range check and error message
	if r.Salary < models.EmpSalaryMin || r.Salary > models.EmpSalaryMax {
		return fmt.Errorf("employee salary must be between %.0f and %.0f", models.EmpSalaryMin, models.EmpSalaryMax)
	}

	return nil
}
