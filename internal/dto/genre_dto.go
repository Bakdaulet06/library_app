package dto

import (
	"errors"
	"strings"
)

// Single DTO for both Creation and Updates
type CreateOrUpdateGenreRequest struct {
	Name string `json:"name"`
}

func (r *CreateOrUpdateGenreRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return errors.New("genre name cannot be empty or blank space")
	}
	return nil
}
