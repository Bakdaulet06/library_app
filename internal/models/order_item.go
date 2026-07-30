package models

// OrderItem represents a single book line within an Order. UnitPrice is
// captured at purchase time (a snapshot of Book.Price), so historical
// orders stay accurate even if the book's price changes later.
type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	BookID    int     `json:"book_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}
