package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/params"
	"library/internal/repositories"
)

var (
	// ErrInsufficientStock is returned when fewer copies are available than requested.
	ErrInsufficientStock = errors.New("not enough copies available at this library")
)

type OrderService interface {
	BuyBooks(ctx context.Context, buyBookRequestFull dto.BuyBookRequestFull) (*BuyBookResult, error)
	GetOrdersByMemberID(ctx context.Context, memberID int) ([]models.Order, error)
	GetOrdersByLibraryID(ctx context.Context, libraryID int, p params.OrderParams) ([]models.Order, error)
	GetAllOrders(ctx context.Context, p params.OrderParams) ([]models.Order, error)
}

type orderService struct {
	db                *sql.DB
	orderRepo         repositories.OrderRepository
	bookInventoryRepo repositories.BookInventoryRepository
	bookshelfRepo     repositories.BookshelfRepository
	profileRepo       repositories.ProfileRepository
}

func NewOrderService(db *sql.DB, orderRepo repositories.OrderRepository, bookInventoryRepo repositories.BookInventoryRepository, bookshelfRepo repositories.BookshelfRepository, profileRepo repositories.ProfileRepository) *orderService {
	return &orderService{db: db, orderRepo: orderRepo, bookInventoryRepo: bookInventoryRepo, bookshelfRepo: bookshelfRepo, profileRepo: profileRepo}
}

// BuyBookResult is the outcome of a successful purchase, handed back to the
// handler layer for JSON serialization.
type BuyBookResult struct {
	OrderID     int
	LibraryID   int
	BookID      int
	Quantity    int
	UnitPrice   float64
	TotalAmount float64
}

// BuyBook purchases `quantity` copies of one book from one library, for one
// member, as a single atomic transaction:
//
//  1. BEGIN
//  2. Lock the book_inventory row (FOR UPDATE) and read available_copies + price
//  3. Verify available_copies >= quantity, else roll back with ErrInsufficientStock
//  4. Decrement available_copies
//  5. Insert the order header + one order_item line
//  6. COMMIT (or ROLLBACK on any error along the way)
//
// The FOR UPDATE lock in step 2 is what makes this safe under concurrency -
// a second simultaneous purchase of the same book blocks at step 2 until
// this transaction commits or rolls back, so it never reads a stale count.
func (s *orderService) BuyBooks(ctx context.Context, req dto.BuyBookRequestFull) (*BuyBookResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch shelf locations inside the transaction
	shelfAllocations, err := s.bookInventoryRepo.GetShelfAllocationsByBook(ctx, tx, req.LibraryID, req.BookID)
	if err != nil {
		return nil, fmt.Errorf("failed to check book shelf locations: %w", err)
	}
	if len(shelfAllocations) == 0 {
		return nil, errors.New("cannot process purchase: book record does not exist for this library")
	}

	// 2. Validate total inventory across all shelves
	var totalAvailable int
	for _, alloc := range shelfAllocations {
		totalAvailable += alloc.Quantity
	}

	if totalAvailable < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// 3. Get book price
	bookLocation := models.BookLocation{
		LibraryID:   req.LibraryID,
		BookID:      req.BookID,
		BookshelfID: shelfAllocations[0].BookshelfID,
	}
	price, err := s.orderRepo.GetBookPrice(ctx, tx, bookLocation)
	if err != nil {
		return nil, err
	}

	// 4. Single Loop: Update both Inventory and Bookshelf Space per shelf
	remainingNeeded := req.Quantity
	for _, alloc := range shelfAllocations {
		if remainingNeeded <= 0 {
			break
		}

		deductAmount := alloc.Quantity
		if deductAmount > remainingNeeded {
			deductAmount = remainingNeeded
		}

		// A. Decrement book stock on this shelf (pass negative deductAmount)
		if err := s.bookInventoryRepo.UpdateAvailableCopies(ctx, tx, req.LibraryID, req.BookID, alloc.BookshelfID, -deductAmount); err != nil {
			return nil, fmt.Errorf("failed to deduct inventory: %w", err)
		}

		// B. Increase empty space on this bookshelf (pass negative deductAmount so SQL does `empty_space - (-deductAmount)`)
		if err := s.bookshelfRepo.UpdateEmptySpace(ctx, tx, req.LibraryID, alloc.BookshelfID, -deductAmount); err != nil {
			return nil, fmt.Errorf("failed to update shelf space: %w", err)
		}

		remainingNeeded -= deductAmount
	}

	// Sanity check
	if remainingNeeded > 0 {
		return nil, fmt.Errorf("could not fulfill order: %d copies unfulfilled", remainingNeeded)
	}

	// 5. Create Order & Order Items
	totalAmount := price * float64(req.Quantity)

	card, err := s.profileRepo.GetCardByUserID(ctx, s.db, req.MemberID)
	if err != nil {
		return nil, err
	}
	if card.Amount < totalAmount {
		return nil, fmt.Errorf("could not fulfill order: not enough money")
	}

	if err := s.profileRepo.UpdateCardBalance(ctx, s.db, card.ID, -totalAmount); err != nil {
		return nil, err
	}

	orderID, err := s.orderRepo.CreateOrder(ctx, tx, req.MemberID, req.LibraryID, totalAmount)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	if err := s.orderRepo.AddOrderItem(ctx, tx, orderID, req.BookID, req.Quantity, price, totalAmount); err != nil {
		return nil, fmt.Errorf("create order item: %w", err)
	}

	// 6. Commit Transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &BuyBookResult{
		OrderID:     orderID,
		LibraryID:   req.LibraryID,
		BookID:      req.BookID,
		Quantity:    req.Quantity,
		UnitPrice:   price,
		TotalAmount: totalAmount,
	}, nil
}

func (s *orderService) GetOrdersByMemberID(ctx context.Context, memberID int) ([]models.Order, error) {
	return s.orderRepo.GetOrdersByMemberID(ctx, s.db, memberID)
}

func (s *orderService) GetOrdersByLibraryID(ctx context.Context, libraryID int, p params.OrderParams) ([]models.Order, error) {
	return s.orderRepo.GetOrdersByLibraryID(ctx, s.db, libraryID, p)
}

func (s *orderService) GetAllOrders(ctx context.Context, p params.OrderParams) ([]models.Order, error) {
	return s.orderRepo.GetAllOrders(ctx, s.db, p)
}
