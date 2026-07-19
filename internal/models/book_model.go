package models

import "time"

// Book represents the core domain entity for library inventory.
type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	Isbn            string    `json:"isbn"`
	AvailableCopies int       `json:"available_copies"`
	CreatedAt       time.Time `json:"created_at"`
}
