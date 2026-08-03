package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"library/internal/dto"
	"library/internal/geocoder"
	"library/internal/models"
	"library/internal/params"
	"library/internal/repositories"
)

type LibraryService interface {
	RegisterLibrary(ctx context.Context, req dto.CreateLibraryRequest) (*models.Library, error)
	GetLibraryByID(ctx context.Context, id int) (*models.Library, error)
	GetLibraryLoans(ctx context.Context, id int) ([]models.Loan, error)
	GetLibraryBooks(ctx context.Context, id int, p params.BookParams) ([]models.LibraryBook, error)
	ListLibraries(ctx context.Context, p params.Pagination) ([]models.Library, error)
	UpdateLibrary(ctx context.Context, id int, req dto.CreateLibraryRequest) (*models.Library, error)
	DeleteLibrary(ctx context.Context, id int) error

	BorrowBook(ctx context.Context, memberID, libraryID, bookID, borrowDays int) error
	ReturnBook(ctx context.Context, libraryID, memberID, bookID int) error
	AssignShelf(ctx context.Context, libraryID, bookID int) (*models.Bookshelf, error)
	ListReturnedBooks(ctx context.Context, libraryID int) ([]models.ReturnedBook, error)
}

type libraryService struct {
	db                *sql.DB
	libraryRepo       repositories.LibraryRepository
	bookRepo          repositories.BookRepository
	memberRepo        repositories.MemberRepository
	bookInventoryRepo repositories.BookInventoryRepository
	bookshelfRepo     repositories.BookshelfRepository
	geocoder          geocoder.Geocoder
	profileRepo       repositories.ProfileRepository
}

func NewLibraryService(
	db *sql.DB,
	libraryRepo repositories.LibraryRepository,
	bookRepo repositories.BookRepository,
	memberRepo repositories.MemberRepository,
	bookInventoryRepo repositories.BookInventoryRepository,
	bookshelfRepo repositories.BookshelfRepository,
	geocoder geocoder.Geocoder,
	profileRepo repositories.ProfileRepository,
) LibraryService {
	return &libraryService{
		db:                db,
		libraryRepo:       libraryRepo,
		bookRepo:          bookRepo,
		memberRepo:        memberRepo,
		bookInventoryRepo: bookInventoryRepo,
		bookshelfRepo:     bookshelfRepo,
		geocoder:          geocoder,
		profileRepo:       profileRepo,
	}
}

// 1. RegisterLibrary
func (s *libraryService) RegisterLibrary(ctx context.Context, req dto.CreateLibraryRequest) (*models.Library, error) {
	normalizedAddress, err := s.geocoder.ValidateAndNormalizeAddress(ctx, req.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// 2. Check if a library is already registered at this verified address
	existingLibrary, err := s.libraryRepo.LibraryExists(ctx, s.db, normalizedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check library existence: %w", err)
	}
	if existingLibrary {
		return nil, errors.New("a library is already registered at this verified address")
	}

	library := &models.Library{
		Name:    req.Name,
		Address: normalizedAddress,
	}

	if err := s.libraryRepo.Create(ctx, s.db, library); err != nil {
		return nil, fmt.Errorf("failed to register library: %w", err)
	}

	return library, nil
}

// 2. GetLibraryByID
func (s *libraryService) GetLibraryByID(ctx context.Context, id int) (*models.Library, error) {
	library, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error fetching library: %w", err)
	}
	if library == nil {
		return nil, errors.New("library branch record does not exist")
	}

	return library, nil
}

// 3. ListLibraries
func (s *libraryService) ListLibraries(ctx context.Context, p params.Pagination) ([]models.Library, error) {
	libraries, err := s.libraryRepo.ListAll(ctx, s.db, p)
	if err != nil {
		return nil, fmt.Errorf("failed to list libraries: %w", err)
	}

	if libraries == nil {
		return []models.Library{}, nil // Prevent returning null in HTTP JSON responses
	}

	return libraries, nil
}

// 4. UpdateLibrary
func (s *libraryService) UpdateLibrary(ctx context.Context, id int, req dto.CreateLibraryRequest) (*models.Library, error) {
	// Verify library exists before updating
	existing, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if existing == nil {
		return nil, errors.New("cannot update: library branch record does not exist")
	}

	library := &models.Library{
		ID:      id,
		Name:    req.Name,
		Address: req.Address,
	}

	if err := s.libraryRepo.Update(ctx, s.db, library); err != nil {
		return nil, fmt.Errorf("failed to update library: %w", err)
	}

	return library, nil
}

// 5. DeleteLibrary
func (s *libraryService) DeleteLibrary(ctx context.Context, id int) error {
	// Verify library exists before attempting delete
	existing, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return fmt.Errorf("database error checking library before delete: %w", err)
	}
	if existing == nil {
		return errors.New("cannot delete: library branch record does not exist")
	}

	if err := s.libraryRepo.Delete(ctx, s.db, id); err != nil {
		return fmt.Errorf("failed to delete library branch: %w", err)
	}

	return nil
}

// 6. GetLibraryLoans
func (s *libraryService) GetLibraryLoans(ctx context.Context, id int) ([]models.Loan, error) {
	// 1. Verify library branch exists
	library, err := s.libraryRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if library == nil {
		return nil, errors.New("cannot fetch loans: targeted library branch does not exist")
	}

	// 2. Fetch loans originated at this branch
	loans, err := s.libraryRepo.ListAllLoans(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library loans: %w", err)
	}

	if loans == nil {
		return []models.Loan{}, nil
	}

	return loans, nil
}

// 7. GetLibraryBooks
func (s *libraryService) GetLibraryBooks(ctx context.Context, libraryID int, p params.BookParams) ([]models.LibraryBook, error) {
	// 1. Verify library branch exists
	library, err := s.libraryRepo.GetByID(ctx, s.db, libraryID)
	if err != nil {
		return nil, fmt.Errorf("database error checking library: %w", err)
	}
	if library == nil {
		return nil, errors.New("cannot fetch books: targeted library branch does not exist")
	}

	// 2. Fetch books stocked at this branch
	books, err := s.libraryRepo.ListAllBooks(ctx, s.db, libraryID, p)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library books: %w", err)
	}

	if books == nil {
		return []models.LibraryBook{}, nil
	}

	return books, nil
}

// --- 1. BorrowBook ---
func (s *libraryService) BorrowBook(ctx context.Context, memberID, libraryID, bookID, borrowDays int) error {
	if err := models.ValidateLoanDays(borrowDays); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Verify Member Exists
	member, err := s.memberRepo.GetByID(ctx, tx, memberID)
	if err != nil {
		return fmt.Errorf("database error checking member: %w", err)
	}
	if member == nil {
		return errors.New("checkout blocked: targeted member record does not exist")
	}

	// 2. Verify member has a card and it's not already in the red
	card, err := s.profileRepo.GetCardByUserID(ctx, tx, memberID)
	if err != nil {
		return fmt.Errorf("checkout blocked: could not verify member card: %w", err)
	}
	if card.Amount < 0 {
		return errors.New("checkout blocked: member's card balance is negative")
	}

	// 3. Verify Book Exists in shelves of library
	shelfAllocations, err := s.bookInventoryRepo.GetShelfAllocationsByBook(ctx, tx, libraryID, bookID)
	if err != nil {
		return fmt.Errorf("failed to check book shelf locations: %w", err)
	}
	if len(shelfAllocations) == 0 {
		return errors.New("there's no such book in the library")
	}

	// 4. Verify Active Loan
	hasLoan, err := s.bookRepo.HasActiveLoan(ctx, tx, bookID, memberID)
	if err != nil {
		return fmt.Errorf("database error checking active loans: %w", err)
	}
	if hasLoan {
		return errors.New("this member has already borrowed this book and has not returned it yet")
	}

	bookLocation := models.BookLocation{
		BookID:      bookID,
		LibraryID:   libraryID,
		BookshelfID: shelfAllocations[0].BookshelfID,
	}

	// 5. Decrement Stock
	if err := s.bookInventoryRepo.DecrementInventory(ctx, tx, bookLocation); err != nil {
		return fmt.Errorf("failed to decrement inventory: %w", err)
	}

	if err := s.bookshelfRepo.UpdateEmptySpace(ctx, tx, libraryID, bookLocation.BookshelfID, -1); err != nil {
		return fmt.Errorf("failed to update empty space: %w", err)
	}

	// 6. Register Loan Record
	if err := s.bookRepo.CreateLoan(ctx, tx, bookID, memberID, libraryID, borrowDays); err != nil {
		return fmt.Errorf("failed to register loan log: %w", err)
	}

	return tx.Commit()
}

// --- 2. ReturnBook ---
// when returning i want to check if user returned their book on time, if they did then we do nothing, but if they returned their book late
// then we have to fine them probably like 1/40 part of the price of the book for one day that wasn't returned, we can fine them
// even if their card has not enough money, their card will go negative,
func (s *libraryService) ReturnBook(ctx context.Context, libraryID, memberID, bookID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	book, err := s.bookRepo.GetByID(ctx, tx, bookID)
	if err != nil {
		return fmt.Errorf("database error checking book catalog: %w", err)
	}
	if book == nil {
		return errors.New("return blocked: target catalog book no longer exists")
	}

	loan, err := s.bookRepo.GetActiveLoan(ctx, tx, bookID, memberID)
	if err != nil {
		return fmt.Errorf("database error checking active loan: %w", err)
	}
	if loan == nil {
		return errors.New("return blocked: no active loan found for this member and book")
	}

	returnedAt := time.Now()

	if err := s.bookRepo.UpdateLoanReturn(ctx, tx, bookID, memberID, libraryID); err != nil {
		return fmt.Errorf("failed to update loan record: %w", err)
	}

	// Fine the member if the book came back late. Capped at the book's price.
	if fine := loan.CalculateFine(returnedAt, book.Price); fine > 0 {
		card, err := s.profileRepo.GetCardByUserID(ctx, tx, memberID)
		if err != nil {
			return fmt.Errorf("failed to fetch card for fine: %w", err)
		}
		if err := s.profileRepo.UpdateCardBalance(ctx, tx, card.ID, -fine); err != nil {
			return fmt.Errorf("failed to apply late fine: %w", err)
		}
	}

	// Queue for employee shelf assignment instead of incrementing inventory directly.
	if err := s.bookRepo.CreateReturnedBook(ctx, tx, bookID, libraryID, memberID); err != nil {
		return fmt.Errorf("failed to queue returned book: %w", err)
	}

	return tx.Commit()
}

// ListReturnedBooks — for GET /libraries/:id/returned_books, employee-only.
func (s *libraryService) ListReturnedBooks(ctx context.Context, libraryID int) ([]models.ReturnedBook, error) {
	return s.bookRepo.GetReturnedBooksByLibrary(ctx, s.db, libraryID)
}

// AssignShelf — for POST /libraries/:id/returned_books/:book_id/assign_shelf, employee-only.
func (s *libraryService) AssignShelf(ctx context.Context, libraryID, bookID int) (*models.Bookshelf, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	returned, err := s.bookRepo.GetReturnedBook(ctx, tx, libraryID, bookID)
	if err != nil {
		return nil, err
	}
	if returned == nil {
		return nil, errors.New("no pending returned record found for this book in this library")
	}

	shelf, err := s.bookshelfRepo.FindAvailableShelf(ctx, tx, libraryID)
	if err != nil {
		return nil, err
	}
	if shelf == nil {
		return nil, errors.New("no available shelf space in this library")
	}

	if err := s.bookshelfRepo.DecrementEmptySpace(ctx, tx, libraryID, shelf.ID); err != nil {
		return nil, err
	}

	bookLocation := &models.BookLocation{
		BookID:      bookID,
		LibraryID:   libraryID,
		BookshelfID: shelf.ID,
	}
	// Now that a shelf is assigned, inventory actually increments.
	if err := s.bookInventoryRepo.IncrementInventory(ctx, tx, *bookLocation); err != nil {
		return nil, fmt.Errorf("failed to increment inventory: %w", err)
	}

	if err := s.bookRepo.DeleteReturnedBook(ctx, tx, returned.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return shelf, nil
}
