package dto

import (
	"errors"
	"strings"
)

// CreateBookshelfRequest represents the JSON payload for creating a new bookshelf.
type CreateBookshelfRequest struct {
	Code     string `json:"code"`
	Capacity int    `json:"capacity"`
}

// Validate ensures bookshelf attributes adhere to operational layout rules.
func (r *CreateBookshelfRequest) Validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return errors.New("bookshelf code parameter cannot be empty")
	}
	if r.Capacity <= 0 {
		return errors.New("bookshelf capacity parameter must be greater than zero")
	}
	return nil
}

// BookshelfResponse defines the standard JSON response format for bookshelf details.
type BookshelfResponse struct {
	ID         int    `json:"id"`
	LibraryID  int    `json:"library_id"`
	Code       string `json:"code"`
	Capacity   int    `json:"capacity"`
	EmptySpace int    `json:"empty_space"`
}

type BookWithShelfStockResponse struct {
	BookID          int    `json:"book_id"`
	Title           string `json:"title"`
	Author          string `json:"author"`
	ISBN            string `json:"isbn"`
	GenreID         string `json:"genre_id"`
	AvailableCopies int    `json:"available_copies"`
}
