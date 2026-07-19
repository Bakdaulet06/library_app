package dto

import (
	"errors"
	"strings"
)

// CreateMemberRequest represents the JSON payload for registering a new user.
type CreateMemberRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Validate performs basic checks for missing fields and simple email formatting.
func (req *CreateMemberRequest) Validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		return errors.New("email is required")
	}
	// Pure standard library basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("invalid email address format")
	}

	return nil
}
