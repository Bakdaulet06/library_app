package models

import "time"

// Loan represents the transaction record when a member borrows a book.
type Loan struct {
	ID         int64      `json:"id"`
	BookID     int64      `json:"book_id"`
	MemberID   int64      `json:"member_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"` // Pointer handles nullable DB timestamps cleanly
}
