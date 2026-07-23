package models

import "time"

// Book represents the core domain entity for library inventory.
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Isbn      string    `json:"isbn"`
	GenreID   int       `json:"genre_id"` // <-- Foreign Key to libraries.id
	CreatedAt time.Time `json:"created_at"`
}
