package models

type BookInventory struct {
	BookLocation
	AvailableCopies int `json:"available_copies"`
}
