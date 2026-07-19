package models

import "time"

// Loan represents the transaction record when a member borrows a book.
type Loan struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	MemberID   int        `json:"member_id"`
	LoanDate   time.Time  `json:"loan_date"`
	ReturnDate *time.Time `json:"return_date"` // Nullable to track active vs completed loans
}
