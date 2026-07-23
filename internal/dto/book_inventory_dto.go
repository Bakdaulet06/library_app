package dto

import (
	"errors"
)

// CreateMemberRequest represents the JSON payload for registering a new user.
type CreateBookInventoryRequest struct {
	LibraryID       int `json:"library_id"`
	BookID          int `json:"book_id"`
	AvailableCopies int `json:"available_copies"`
}

// Validate ensures profile attributes adhere to operational layout rules.
func (r *CreateBookInventoryRequest) Validate() error {
	if r.BookID <= 0 {
		return errors.New("invalid target catalog book_id parameter")
	}
	if r.LibraryID <= 0 {
		return errors.New("invalid processing library_id parameter")
	}
	if r.AvailableCopies <= 0 {
		return errors.New("invalid processing available_copies parameter")
	}
	return nil
}
