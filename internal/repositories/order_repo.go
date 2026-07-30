package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/models"
)

// ErrBookNotInLibrary is returned when the library/book pair has no
// inventory row at all (as opposed to zero copies available).
var ErrBookNotInLibrary = errors.New("book is not stocked at this library")

type OrderRepository interface {
	GetBookPrice(ctx context.Context, tx *sql.Tx, bookLocation models.BookLocation) (price float64, err error)
	CreateOrder(ctx context.Context, tx *sql.Tx, memberID, libraryID int, totalAmount float64) (int, error)
	AddOrderItem(ctx context.Context, tx *sql.Tx, orderID, bookID, quantity int, unitPrice, subtotal float64) error
	GetOrdersByMemberID(ctx context.Context, exec GormExecutor, memberID int) ([]models.Order, error)
	GetOrdersByLibraryID(ctx context.Context, exec GormExecutor, libraryID int) ([]models.Order, error)
	GetAllOrders(ctx context.Context, exec GormExecutor) ([]models.Order, error)
}

type orderRepository struct{}

func NewOrderRepository() *orderRepository {
	return &orderRepository{}
}

// LockInventoryAndPrice reads the current available_copies and the book's
// price, taking a row lock (FOR UPDATE) that's held until the caller's
// transaction commits or rolls back. This is what prevents two concurrent
// purchases from both reading "1 copy left" and both succeeding (oversell).
//
// Must be called inside a transaction - the lock is meaningless outside one.
func (r *orderRepository) GetBookPrice(ctx context.Context, tx *sql.Tx, bookLocation models.BookLocation) (price float64, err error) {
	const query = `
		SELECT b.price
		FROM book_inventory bi
		JOIN books b ON b.id = bi.book_id
		WHERE bi.library_id = $1 AND bi.book_id = $2 AND bi.bookshelf_id = $3
		FOR UPDATE
	`
	row := tx.QueryRowContext(ctx, query, bookLocation.LibraryID, bookLocation.BookID, bookLocation.BookshelfID)
	if err := row.Scan(&price); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrBookNotInLibrary
		}
		return 0, err
	}
	return price, nil
}

// CreateOrder inserts the order header row and returns its generated ID.
func (r *orderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, memberID, libraryID int, totalAmount float64) (int, error) {
	const query = `
		INSERT INTO orders (member_id, library_id, total_amount, created_at)
		VALUES ($1, $2, $3, now())
		RETURNING id
	`
	var orderID int
	err := tx.QueryRowContext(ctx, query, memberID, libraryID, totalAmount).Scan(&orderID)
	return orderID, err
}

// AddOrderItem inserts a single line item for an order.
func (r *orderRepository) AddOrderItem(ctx context.Context, tx *sql.Tx, orderID, bookID, quantity int, unitPrice, subtotal float64) error {
	const query = `
		INSERT INTO order_items (order_id, book_id, quantity, unit_price, subtotal)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(ctx, query, orderID, bookID, quantity, unitPrice, subtotal)
	return err
}

// GetOrdersByMemberID fetches all orders along with their nested line items for a given member
func (r *orderRepository) GetOrdersByMemberID(ctx context.Context, exec GormExecutor, memberID int) ([]models.Order, error) {
	// 1. Fetch parent orders
	orderQuery := `
		SELECT id, member_id, library_id, total_amount, created_at
		FROM orders
		WHERE member_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.QueryContext(ctx, orderQuery, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders for member: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	orderMap := make(map[int]*models.Order)

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.MemberID, &o.LibraryID, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = []models.OrderItem{} // initialize empty slice
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return []models.Order{}, nil
	}

	// Store pointers in map to easily attach items
	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
	}

	// 2. Fetch corresponding order_items via JOIN
	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.book_id, oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.member_id = $1
		ORDER BY oi.id ASC`

	itemRows, err := exec.QueryContext(ctx, itemsQuery, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.OrderItem
		if err := itemRows.Scan(&item.ID, &item.OrderID, &item.BookID, &item.Quantity, &item.UnitPrice, &item.Subtotal); err != nil {
			return nil, err
		}
		if parentOrder, exists := orderMap[item.OrderID]; exists {
			parentOrder.Items = append(parentOrder.Items, item)
		}
	}

	return orders, nil
}

// GetOrdersByLibraryID fetches all orders along with their nested line items for a given library
func (r *orderRepository) GetOrdersByLibraryID(ctx context.Context, exec GormExecutor, libraryID int) ([]models.Order, error) {
	// 1. Fetch parent orders
	orderQuery := `
		SELECT id, member_id, library_id, total_amount, created_at
		FROM orders
		WHERE library_id = $1
		ORDER BY created_at DESC`

	rows, err := exec.QueryContext(ctx, orderQuery, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders for library: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	orderMap := make(map[int]*models.Order)

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.MemberID, &o.LibraryID, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = []models.OrderItem{}
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return []models.Order{}, nil
	}

	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
	}

	// 2. Fetch corresponding order_items
	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.book_id, oi.quantity, oi.unit_price, oi.subtotal
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.library_id = $1
		ORDER BY oi.id ASC`

	itemRows, err := exec.QueryContext(ctx, itemsQuery, libraryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.OrderItem
		if err := itemRows.Scan(&item.ID, &item.OrderID, &item.BookID, &item.Quantity, &item.UnitPrice, &item.Subtotal); err != nil {
			return nil, err
		}
		if parentOrder, exists := orderMap[item.OrderID]; exists {
			parentOrder.Items = append(parentOrder.Items, item)
		}
	}

	return orders, nil
}

// GetAllOrders fetches every order and its line items in the system
func (r *orderRepository) GetAllOrders(ctx context.Context, exec GormExecutor) ([]models.Order, error) {
	orderQuery := `
		SELECT id, member_id, library_id, total_amount, created_at
		FROM orders
		ORDER BY created_at DESC`

	rows, err := exec.QueryContext(ctx, orderQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	orderMap := make(map[int]*models.Order)

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.MemberID, &o.LibraryID, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = []models.OrderItem{}
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return []models.Order{}, nil
	}

	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
	}

	// Fetch all line items attached to these orders
	itemsQuery := `
		SELECT id, order_id, book_id, quantity, unit_price, subtotal
		FROM order_items
		ORDER BY id ASC`

	itemRows, err := exec.QueryContext(ctx, itemsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.OrderItem
		if err := itemRows.Scan(&item.ID, &item.OrderID, &item.BookID, &item.Quantity, &item.UnitPrice, &item.Subtotal); err != nil {
			return nil, err
		}
		if parentOrder, exists := orderMap[item.OrderID]; exists {
			parentOrder.Items = append(parentOrder.Items, item)
		}
	}

	return orders, nil
}
