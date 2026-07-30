package dto

import "errors"

// BuyBookRequest is the payload for POST /libraries/{id}/books/{book_id}/buy.
// The library and book come from the URL path; only quantity is in the body.
type BuyBookRequest struct {
	Quantity int `json:"quantity"`
}

type BuyBookRequestFull struct {
	MemberID  int `json:"member_id"`
	LibraryID int `json:"library_id"`
	BookID    int `json:"book_id"`
	CardID    int `json:"card_id"`
	Quantity  int `json:"quantity"`
}

func (r *BuyBookRequest) Validate() error {
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	return nil
}

// BuyBookResponse is returned after a successful purchase.
type BuyBookResponse struct {
	OrderID     int     `json:"order_id"`
	LibraryID   int     `json:"library_id"`
	BookID      int     `json:"book_id"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalAmount float64 `json:"total_amount"`
}
