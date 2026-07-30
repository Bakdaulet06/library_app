package dto

import (
	"errors"
	"library/internal/models"
	"strings"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // Default to "client" if empty
}

type RegisterClientAndLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string            `json:"token"`
	User  models.BookMember `json:"user"`
}

type UpdateClientProfile struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (r RegisterRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if r.Role != "" && r.Role != "client" && r.Role != "employee" {
		return errors.New("role must be either 'client' or 'employee'")
	}
	return nil
}

func (r RegisterClientAndLoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func (r UpdateClientProfile) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}
	if r.OldPassword == "" || r.NewPassword == "" {
		return errors.New("password can't be empty")
	}
	return nil
}
