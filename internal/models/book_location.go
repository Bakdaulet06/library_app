package models

type BookLocation struct {
	LibraryID   int `json:"library_id"`
	BookID      int `json:"book_id"`
	BookshelfID int `json:"bookshelf_id"`
}
