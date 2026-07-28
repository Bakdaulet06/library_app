package models

import "time"

type ReturnedBook struct {
	ID         int       `json:"id"`
	BookID     int       `json:"book_id"`
	LibraryID  int       `json:"library_id"`
	MemberID   int       `json:"member_id"`
	ReturnedAt time.Time `json:"returned_at"`
}
