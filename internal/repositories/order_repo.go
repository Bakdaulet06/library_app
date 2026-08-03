package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/internal/models"
	"library/internal/params"
	"strconv"
	"strings"
)

// ErrBookNotInLibrary is returned when the library/book pair has no
// inventory row at all (as opposed to zero copies available).
var ErrBookNotInLibrary = errors.New("book is not stocked at this library")

type OrderRepository interface {
	GetBookPrice(ctx context.Context, tx *sql.Tx, bookLocation models.BookLocation) (price float64, err error)
	CreateOrder(ctx context.Context, tx *sql.Tx, memberID, libraryID int, totalAmount float64) (int, error)
	AddOrderItem(ctx context.Context, tx *sql.Tx, orderID, bookID, quantity int, unitPrice, subtotal float64) error
	GetOrdersByMemberID(ctx context.Context, exec GormExecutor, memberID int) ([]models.Order, error)
	GetOrdersByLibraryID(ctx context.Context, exec GormExecutor, libraryID int, p params.OrderParams) ([]models.Order, error)
	GetAllOrders(ctx context.Context, exec GormExecutor, p params.OrderParams) ([]models.Order, error)
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
func (r *orderRepository) GetOrdersByLibraryID(ctx context.Context, exec GormExecutor, libraryID int, p params.OrderParams) ([]models.Order, error) {
	// 1. Build Base Order Query
	orderQuery := `
		SELECT id, member_id, library_id, total_amount, created_at
		FROM orders
		WHERE library_id = $1`

	args := []interface{}{libraryID}
	paramIdx := 2 // $1 is reserved for libraryID

	// Filter by Member ID
	if p.MemberID > 0 {
		orderQuery += fmt.Sprintf(" AND member_id = $%d", paramIdx)
		args = append(args, p.MemberID)
		paramIdx++
	}

	// Optional Numeric ID search (e.g., search specific order_id or member_id)
	if p.Search != "" {
		if idSearch, err := strconv.Atoi(p.Search); err == nil && idSearch > 0 {
			orderQuery += fmt.Sprintf(" AND (id = $%d OR member_id = $%d)", paramIdx, paramIdx+1)
			args = append(args, idSearch, idSearch)
			paramIdx += 2
		}
	}

	// Column Allowlist for Safe Sorting
	allowedColumns := map[string]string{
		"id":           "id",
		"member_id":    "member_id",
		"total_amount": "total_amount",
		"created_at":   "created_at",
	}

	sortColumn, exists := allowedColumns[strings.ToLower(p.SortBy)]
	if !exists {
		sortColumn = "created_at" // Default sort column
	}

	orderDir := strings.ToUpper(p.Order)
	if orderDir != "ASC" {
		orderDir = "DESC" // Default ordering direction for orders
	}

	orderQuery += fmt.Sprintf(" ORDER BY %s %s", sortColumn, orderDir)

	// Apply Limit & Offset
	orderQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, p.Limit, p.Offset)

	// Execute Parent Orders Query
	rows, err := exec.QueryContext(ctx, orderQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders for library: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	orderMap := make(map[int]*models.Order)
	var orderIDs []interface{}

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.MemberID, &o.LibraryID, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = []models.OrderItem{} // Ensure empty array instead of null
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return []models.Order{}, nil
	}

	// Map pointers and gather order IDs for optimized line item fetching
	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
		orderIDs = append(orderIDs, orders[i].ID)
	}

	// 2. Fetch Order Items strictly for the paged orders ($1, $2, ...)
	inPlaceholders := make([]string, len(orderIDs))
	for i := range orderIDs {
		inPlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}

	itemsQuery := fmt.Sprintf(`
		SELECT id, order_id, book_id, quantity, unit_price, subtotal
		FROM order_items
		WHERE order_id IN (%s)
		ORDER BY id ASC`, strings.Join(inPlaceholders, ", "))

	itemRows, err := exec.QueryContext(ctx, itemsQuery, orderIDs...)
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

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetAllOrders fetches every order and its line items in the system
// GetAllOrders fetches every order and its line items in the system with pagination & filters
func (r *orderRepository) GetAllOrders(ctx context.Context, exec GormExecutor, p params.OrderParams) ([]models.Order, error) {
	query := `
		SELECT id, member_id, library_id, total_amount, created_at
		FROM orders`

	var whereClauses []string
	var args []interface{}
	paramIdx := 1

	// 1. Optional Filter by Member ID
	if p.MemberID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("member_id = $%d", paramIdx))
		args = append(args, p.MemberID)
		paramIdx++
	}

	// 2. Optional Numeric Search Filter (Order ID, Member ID, or Library ID)
	if p.Search != "" {
		if idSearch, err := strconv.Atoi(p.Search); err == nil && idSearch > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("(id = $%d OR member_id = $%d OR library_id = $%d)", paramIdx, paramIdx+1, paramIdx+2))
			args = append(args, idSearch, idSearch, idSearch)
			paramIdx += 3
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 3. Allowlist Sorting Columns (Prevents SQL Injection)
	allowedColumns := map[string]string{
		"id":           "id",
		"member_id":    "member_id",
		"library_id":   "library_id",
		"total_amount": "total_amount",
		"created_at":   "created_at",
	}

	sortColumn, exists := allowedColumns[strings.ToLower(p.SortBy)]
	if !exists {
		sortColumn = "created_at" // Default sort column
	}

	orderDir := strings.ToUpper(p.Order)
	if orderDir != "ASC" {
		orderDir = "DESC" // Default ordering direction for orders
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, orderDir)

	// 4. Limit & Offset
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	args = append(args, p.Limit, p.Offset)

	// 5. Execute Parent Orders Query
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	orderMap := make(map[int]*models.Order)
	var orderIDs []interface{}

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.MemberID, &o.LibraryID, &o.TotalAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = []models.OrderItem{} // Ensure empty slice [] instead of null in JSON
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return []models.Order{}, nil
	}

	// Map order pointers and extract IDs for optimized line-item fetching
	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
		orderIDs = append(orderIDs, orders[i].ID)
	}

	// 6. Fetch Order Items strictly for the paged parent orders ($1, $2, ...)
	inPlaceholders := make([]string, len(orderIDs))
	for i := range orderIDs {
		inPlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}

	itemsQuery := fmt.Sprintf(`
		SELECT id, order_id, book_id, quantity, unit_price, subtotal
		FROM order_items
		WHERE order_id IN (%s)
		ORDER BY id ASC`, strings.Join(inPlaceholders, ", "))

	itemRows, err := exec.QueryContext(ctx, itemsQuery, orderIDs...)
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

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
