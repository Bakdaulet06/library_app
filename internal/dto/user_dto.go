package dto

import (
	"errors"
	"library/internal/models"
	"net/mail"
	"strings"
)

type RegisterRequest struct {
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

// Helper function to validate email standard format
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(strings.TrimSpace(email))
	return err == nil
}

func (r RegisterRequest) Validate() error {
	if !isValidEmail(r.Email) {
		return errors.New("invalid email address")
	}
	if len(r.Password) <= 8 {
		return errors.New("password must be over 8 characters long")
	}
	return nil
}

func (r UpdateClientProfile) Validate() error {
	if !isValidEmail(r.Email) {
		return errors.New("invalid email address")
	}
	if len(r.OldPassword) <= 8 || len(r.NewPassword) <= 8 {
		return errors.New("passwords must be over 8 characters long")
	}
	return nil
}
