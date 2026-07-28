package dto

import (
	"errors"
	"net/mail"
	"strings"
)

type CreateEmployeeRequest struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	LibraryID int     `json:"library_id"`
	Position  string  `json:"position"`
	Salary    float64 `json:"salary"`
}

func (r *CreateEmployeeRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Position = strings.TrimSpace(r.Position)

	if r.Email == "" {
		return errors.New("email address is required")
	}

	if r.LibraryID <= 0 {
		return errors.New("invalid library ID")
	}

	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email address format provided")
	}

	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	if r.Position == "" {
		return errors.New("employee position title is required")
	}

	if r.Salary < 0 {
		return errors.New("employee salary cannot be negative")
	}

	return nil
}

type UpdateEmployeeRequest struct {
	Position string  `json:"position"`
	Salary   float64 `json:"salary"`
}

func (r *UpdateEmployeeRequest) Validate() error {
	r.Position = strings.TrimSpace(r.Position)

	if r.Position == "" {
		return errors.New("employee position title is required")
	}

	if r.Salary < 0 {
		return errors.New("employee salary cannot be negative")
	}

	return nil
}
