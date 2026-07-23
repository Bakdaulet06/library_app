package models

type BookInventory struct {
	LibraryID       int `json:"library_id"`
	BookID          int `json:"book_id"`
	AvailableCopies int `json:"available_copies"`
}
