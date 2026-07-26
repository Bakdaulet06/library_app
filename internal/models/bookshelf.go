package models

import "time"

type Bookshelf struct {
	ID         int       `json:"id"`
	LibraryID  int       `json:"library_id"`
	Code       string    `json:"code"`
	Capacity   int       `json:"capacity"`
	EmptySpace int       `json:"empty_space"`
	CreatedAt  time.Time `json:"created_at"`
}
