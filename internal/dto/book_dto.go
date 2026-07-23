package dto

import (
	"errors"
	"strings"
)

// CreateBookRequest handles input payloads for both cataloging and updating books.
type CreateBookRequest struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	Isbn            string `json:"isbn"`
	AvailableCopies int    `json:"available_copies"`
	GenreID         *int   `json:"genre_id"`
	LibraryID       int    `json:"library_id"`
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
	if r.AvailableCopies < 0 {
		return errors.New("available copies balance inventory cannot be negative numbers")
	}
	if r.GenreID != nil {
		if *r.GenreID < 1 || *r.GenreID > 5 {
			return errors.New("genre_id must be a positive integer greater than 0 and less than 5")
		}
	} else {
		defaultGenre := 1
		r.GenreID = &defaultGenre
	}
	return nil
}

// CreateMemberRequest processes registrations and profile update actions.

// BorrowBookRequest ingests identification parameters to issue an active asset loan.
type BorrowBookRequest struct {
	BookID              int `json:"book_id"`
	MemberID            int `json:"member_id"`
	Borrowed_library_id int `json:"borrowed_library_id"`
}

// Validate confirms ID allocations are realistic before database processing.
func (r *BorrowBookRequest) Validate() error {
	if r.BookID <= 0 {
		return errors.New("invalid target catalog book_id parameter")
	}
	if r.MemberID <= 0 {
		return errors.New("invalid processing library member_id parameter")
	}
	if r.Borrowed_library_id <= 0 {
		return errors.New("invalid processing library borrowed_library_id parameter")
	}
	return nil
}

// ReturnBookRequest carries indicators necessary to conclude open loan segments.
type ReturnBookRequest struct {
	BookID              int `json:"book_id"`
	MemberID            int `json:"member_id"`
	Returned_library_id int `json:"returned_library_id"`
}

// Validate ensures entity key signatures match expectations.
func (r *ReturnBookRequest) Validate() error {
	if r.BookID <= 0 {
		return errors.New("invalid return target book_id verification token")
	}
	if r.MemberID <= 0 {
		return errors.New("invalid return member processing member_id identification token")
	}
	if r.Returned_library_id <= 0 {
		return errors.New("invalid processing library returned_library_id parameter")
	}
	return nil
}
