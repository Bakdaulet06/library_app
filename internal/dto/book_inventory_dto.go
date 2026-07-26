package dto

import (
	"errors"
)

// CreateBookInventoryRequest represents the JSON payload for adding book inventory to a shelf.
type CreateBookInventoryRequest struct {
	LibraryID       int `json:"library_id"`
	BookID          int `json:"book_id"`
	BookshelfID     int `json:"bookshelf_id"`
	AvailableCopies int `json:"available_copies"`
}

// Validate ensures inventory attributes adhere to operational layout rules.
func (r *CreateBookInventoryRequest) Validate() error {
	if r.BookID <= 0 {
		return errors.New("invalid target catalog book_id parameter")
	}
	if r.LibraryID <= 0 {
		return errors.New("invalid processing library_id parameter")
	}
	if r.BookshelfID <= 0 {
		return errors.New("invalid processing bookshelf_id parameter")
	}
	if r.AvailableCopies <= 0 {
		return errors.New("invalid processing available_copies parameter")
	}
	return nil
}
