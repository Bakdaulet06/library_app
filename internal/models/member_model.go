package models

import "time"

// Member represents the core domain entity for a registered library user.
type BookMember struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Password string    `json:"password"`
	JoinedAt time.Time `json:"joined_at"`
}
