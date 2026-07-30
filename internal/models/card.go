package models

import "time"

// Card represents a user's financial/membership card linked to their member account.
type Card struct {
	ID        int       `json:"id"`
	MemberID  int       `json:"member_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
