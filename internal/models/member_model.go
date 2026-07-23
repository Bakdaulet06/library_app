package models

import "time"

// Member represents the core domain entity for a registered library user.
type BookMember struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	JoinedAt time.Time `json:"joined_at"`
}
