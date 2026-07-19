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

// Validate ensures profile attributes adhere to operational layout rules.
func (r *CreateMemberRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("member name entry cannot be blank")
	}
	emailClean := strings.TrimSpace(r.Email)
	if emailClean == "" || !strings.Contains(emailClean, "@") || !strings.Contains(emailClean, ".") {
		return errors.New("a valid structural email address mapping is required")
	}
	return nil
}
