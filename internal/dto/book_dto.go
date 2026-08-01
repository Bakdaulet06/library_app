package dto

import (
	"errors"
	"strings"
)

// CreateBookRequest handles input payloads for both cataloging and updating books.
type CreateBookRequest struct {
	Title   string  `json:"title"`
	Author  string  `json:"author"`
	Isbn    string  `json:"isbn"`
	GenreID *int    `json:"genre_id,omitempty"`
	Price   float64 `json:"price"` // Can be omitted or set to null by client
}

// Validate asserts data sanity for book creations and modifications.
func (r *CreateBookRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title field cannot be blank or empty space")
	}
	if strings.TrimSpace(r.Author) == "" {
		return errors.New("author field cannot be blank or empty space")
	}
	isbnClean := strings.TrimSpace(r.Isbn)
	if len(isbnClean) != 10 && len(isbnClean) != 13 {
		return errors.New("isbn code must be exactly 10 or 13 characters long")
	}

	// Price Validation
	if r.Price < 0 || r.Price > 100000 {
		return errors.New("price has to be in range of 0 to 100000")
	}
	if r.GenreID != nil && *r.GenreID <= 0 {
		return errors.New("genre_id must be a positive integer greater than 0")
	}

	return nil
}

// CreateMemberRequest processes registrations and profile update actions.

type ReturnBookRequest struct {
	BookID int `json:"book_id"`
}

// Validate ensures entity key signatures match expectations.
func (r *ReturnBookRequest) Validate() error {
	if r.BookID <= 0 {
		return errors.New("invalid return target book_id verification token")
	}
	return nil
}
