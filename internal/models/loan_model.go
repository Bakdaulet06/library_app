package models

import "time"

// Loan represents the transaction record when a member borrows a book.
type Loan struct {
	ID                int        `json:"id"`
	BookID            int        `json:"book_id"`
	MemberID          int        `json:"member_id"`
	BorrowedLibraryID int        `json:"borrowed_library_id"`
	ReturnedLibraryID *int       `json:"returned_library_id,omitempty"`
	BorrowedAt        time.Time  `json:"borrowed_at"`
	ReturnedAt        *time.Time `json:"returned_at,omitempty"` // Pointer handles nullable DB timestamps cleanly
}
