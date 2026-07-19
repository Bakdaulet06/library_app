package dto

import (
	"errors"
	"strings"
)

// CreateBookRequest represents the JSON payload for adding a new book.
type CreateBookRequest struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	ISBN            string `json:"isbn"`
	AvailableCopies int    `json:"available_copies"`
}

// Validate ensures the incoming book data is complete and mathematically sound.
func (req *CreateBookRequest) Validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(req.Author) == "" {
		return errors.New("author is required")
	}

	isbn := strings.TrimSpace(req.ISBN)
	// Basic ISBN sanity check (ISBN-10 or ISBN-13)
	if len(isbn) != 10 && len(isbn) != 13 {
		return errors.New("isbn must be exactly 10 or 13 characters")
	}

	if req.AvailableCopies < 0 {
		return errors.New("available_copies cannot be negative")
	}
	return nil
}

// BorrowBookRequest represents the JSON payload to check out a book.
type BorrowBookRequest struct {
	BookID   int `json:"book_id"`
	MemberID int `json:"member_id"`
}

// Validate ensures positive IDs are provided.
func (req *BorrowBookRequest) Validate() error {
	if req.BookID <= 0 {
		return errors.New("valid book_id is required")
	}
	if req.MemberID <= 0 {
		return errors.New("valid member_id is required")
	}
	return nil
}

// ReturnBookRequest represents the JSON payload to return a borrowed book.
// While identical structurally to BorrowBookRequest, keeping it distinct
// ensures our code represents our domain terminology accurately.
type ReturnBookRequest struct {
	BookID   int `json:"book_id"`
	MemberID int `json:"member_id"`
}

// Validate ensures positive IDs are provided.
func (req *ReturnBookRequest) Validate() error {
	if req.BookID <= 0 {
		return errors.New("valid book_id is required")
	}
	if req.MemberID <= 0 {
		return errors.New("valid member_id is required")
	}
	return nil
}
