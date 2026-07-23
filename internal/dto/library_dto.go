package dto

import (
	"errors"
	"strings"
)

// CreateMemberRequest represents the JSON payload for registering a new user.
type CreateLibraryRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Validate ensures profile attributes adhere to operational layout rules.
func (r *CreateLibraryRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("library name entry cannot be blank")
	}
	if strings.TrimSpace(r.Address) == "" {
		return errors.New("address is required")
	}
	return nil
}
