package dto

import (
	"errors"
	"fmt"
)

// CreateBookshelfRequest represents the JSON payload for creating a new bookshelf.
type CreateBookshelfRequest struct {
	Capacity int `json:"capacity"`
}

// Validate ensures bookshelf attributes adhere to operational layout rules.
func (r *CreateBookshelfRequest) Validate() error {
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

func GenerateBookshelfCode(index int) (string, error) {
	// 26 letters (A-Z) * 10 digits (0-9) = 260 total valid codes
	const maxCodes = 26 * 10

	if index < 0 || index >= maxCodes {
		return "", errors.New("can't create more bookshelves")
	}

	// Calculate character letter ('A' to 'Z')
	letter := rune('A' + (index / 10))

	// Calculate digit number (0 to 9)
	digit := index % 10

	return fmt.Sprintf("%c-%d", letter, digit), nil
}

func ParseCodeToIndex(code string) (int, error) {
	if len(code) != 3 || code[1] != '-' {
		return 0, fmt.Errorf("invalid code format: %s", code)
	}

	letter := code[0]
	digit := code[2]

	if letter < 'A' || letter > 'Z' || digit < '0' || digit > '9' {
		return 0, fmt.Errorf("invalid code characters: %s", code)
	}

	letterIndex := int(letter - 'A')
	digitIndex := int(digit - '0')

	return (letterIndex * 10) + digitIndex, nil
}
