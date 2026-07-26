package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`    // Never expose password in JSON
	Role      string    `json:"role"` // "admin", "employee", "user"
	CreatedAt time.Time `json:"created_at"`
}
