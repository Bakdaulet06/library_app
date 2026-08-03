package models

import (
	"fmt"
	"time"
)

const (
	MinLoanDays = 7
	MaxLoanDays = 21
)

// Loan represents the transaction record when a member borrows a book.
type Loan struct {
	ID                int        `json:"id"`
	BookID            int        `json:"book_id"`
	MemberID          int        `json:"member_id"`
	BorrowedLibraryID int        `json:"borrowed_library_id"`
	ReturnedLibraryID *int       `json:"returned_library_id,omitempty"`
	BorrowedAt        time.Time  `json:"borrowed_at"`
	BorrowedDays      int        `json:"borrowed_days"`
	ReturnedAt        *time.Time `json:"returned_at,omitempty"`
}

// ValidateLoanDays checks that the requested borrow period is within the allowed range.
func ValidateLoanDays(days int) error {
	if days < MinLoanDays || days > MaxLoanDays {
		return fmt.Errorf("borrow period must be between 7 and 21 days %d", days)
	}
	return nil
}

// DueAt returns the deadline by which the book should be returned.
func (l *Loan) DueAt() time.Time {
	return l.BorrowedAt.AddDate(0, 0, l.BorrowedDays)
}

// DaysLate returns how many whole days overdue the loan is as of `at`.
// Returns 0 if it isn't overdue yet.
func (l *Loan) DaysLate(at time.Time) int {
	due := l.DueAt()
	if !at.After(due) {
		return 0
	}
	late := at.Sub(due)
	days := int(late / (24 * time.Hour))
	if late%(24*time.Hour) > 0 {
		days++
	}
	return days
}

// DaysRemaining returns days left before the due date, as of `at`.
// Positive = days left, 0 = due today, negative = overdue by that many days.
// Always the exact negative of DaysLate.
func (l *Loan) DaysRemaining(at time.Time) int {
	return -l.DaysLate(at)
}

// IsOverdue is a convenience check.
func (l *Loan) IsOverdue(at time.Time) bool {
	return l.DaysLate(at) > 0
}

// CalculateFine returns the late fee (1/40th of book price per day late),
// capped so it never exceeds the book's price.
func (l *Loan) CalculateFine(at time.Time, bookPrice float64) float64 {
	daysLate := l.DaysLate(at)
	if daysLate == 0 {
		return 0
	}
	fine := bookPrice / 40 * float64(daysLate)
	if fine > bookPrice {
		return bookPrice
	}
	return fine
}
