package models

type LibraryBook struct {
	Book                 Book
	TotalAvailableCopies int `json:"total_available_copies"`
}
