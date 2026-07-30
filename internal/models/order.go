package models

import "time"

// Order represents a single purchase transaction, scoped to one library
// (books are decremented from that library's inventory).
type Order struct {
	ID          int         `json:"id"`
	MemberID    int         `json:"member_id"`
	LibraryID   int         `json:"library_id"`
	TotalAmount float64     `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
	Items       []OrderItem `json:"items,omitempty"` // attached on JOINs / detail fetches
}
