package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/services"
)

type BookshelfHandler struct {
	bookshelfService services.BookshelfService
}

func NewBookshelfHandler(bookshelfService services.BookshelfService) *BookshelfHandler {
	return &BookshelfHandler{
		bookshelfService: bookshelfService,
	}
}

func (h *BookshelfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Strip static prefix "/libraries/"
	// Example: "/libraries/5/bookshelves"    -> "5/bookshelves"
	// Example: "/libraries/5/bookshelves/12" -> "5/bookshelves/12"
	path := strings.TrimPrefix(r.URL.Path, "/libraries/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Validate path structure: must be ["<library_id>", "bookshelves", ...]
	if len(parts) < 2 || parts[1] != "bookshelves" {
		http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
		return
	}

	// 2. Parse library_id from parts[0]
	libraryID, err := strconv.Atoi(parts[0])
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library_id parameter"}`, http.StatusBadRequest)
		return
	}

	// 3. Route based on URL structure and HTTP method
	switch len(parts) {

	// Route: /libraries/{library_id}/bookshelves
	case 2:
		switch r.Method {
		case http.MethodGet:
			h.GetBookshelvesByLibraryID(w, r, libraryID)
			return
		case http.MethodPost:
			h.CreateBookshelf(w, r, libraryID)
			return
		}

	// Route: /libraries/{library_id}/bookshelves/{shelf_id}
	case 3:
		shelfID, err := strconv.Atoi(parts[2])
		if err != nil || shelfID <= 0 {
			http.Error(w, `{"error":"invalid bookshelf id parameter"}`, http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.GetBookshelfByID(w, r, libraryID, shelfID)
			return
		case http.MethodDelete:
			h.DeleteBookshelf(w, r, libraryID, shelfID)
			return
		}

	// /libraries/{library_id}/bookshelves/{shelf_id}/books
	case 4:
		shelfID, err := strconv.Atoi(parts[2])
		if err != nil || shelfID <= 0 {
			http.Error(w, `{"error":"invalid bookshelf id parameter"}`, http.StatusBadRequest)
			return
		}
		if parts[3] != "books" {
			http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetBooksByShelfID(w, r, libraryID, shelfID)
			return
		}

	// /libraries/{library_id}/bookshelves/{shelf_id}/books/{books_id}
	case 5:
		shelfID, err := strconv.Atoi(parts[2])
		if err != nil || shelfID <= 0 {
			http.Error(w, `{"error":"invalid bookshelf id parameter"}`, http.StatusBadRequest)
			return
		}
		if parts[3] != "books" {
			http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
			return
		}
		bookID, err := strconv.Atoi(parts[4])
		if err != nil || bookID <= 0 {
			http.Error(w, `{"error":"invalid book id parameter"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetBookByShelfID(w, r, libraryID, shelfID, bookID)
			return
		}

	// /libraries/{library_id}/bookshelves/{shelf_id}/books/{books_id}/borrow
	case 6:
		shelfID, err := strconv.Atoi(parts[2])
		if err != nil || shelfID <= 0 {
			http.Error(w, `{"error":"invalid bookshelf id parameter"}`, http.StatusBadRequest)
			return
		}
		if parts[3] != "books" && parts[5] != "borrow" {
			http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
			return
		}
		bookID, err := strconv.Atoi(parts[4])
		if err != nil || bookID <= 0 {
			http.Error(w, `{"error":"invalid book id parameter"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPost:
			h.GetBookByShelfID(w, r, libraryID, shelfID, bookID)
			return
		}
	}

	http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
}

// POST /libraries/{library_id}/bookshelves
func (h *BookshelfHandler) CreateBookshelf(w http.ResponseWriter, r *http.Request, libraryID int) {
	var req dto.CreateBookshelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	shelf := &models.Bookshelf{
		LibraryID: libraryID,
		Code:      req.Code,
		Capacity:  req.Capacity,
	}

	if err := h.bookshelfService.CreateBookshelf(r.Context(), shelf); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	resp := dto.BookshelfResponse{
		ID:         shelf.ID,
		LibraryID:  shelf.LibraryID,
		Code:       shelf.Code,
		Capacity:   shelf.Capacity,
		EmptySpace: shelf.EmptySpace,
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /libraries/{library_id}/bookshelves
func (h *BookshelfHandler) GetBookshelvesByLibraryID(w http.ResponseWriter, r *http.Request, libraryID int) {
	shelves, err := h.bookshelfService.GetBookshelvesByLibraryID(r.Context(), libraryID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	responseList := make([]dto.BookshelfResponse, 0, len(shelves))
	for _, shelf := range shelves {
		responseList = append(responseList, dto.BookshelfResponse{
			ID:         shelf.ID,
			LibraryID:  shelf.LibraryID,
			Code:       shelf.Code,
			Capacity:   shelf.Capacity,
			EmptySpace: shelf.EmptySpace,
		})
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responseList)
}

// GET /libraries/{library_id}/bookshelves/{shelf_id}
func (h *BookshelfHandler) GetBookshelfByID(w http.ResponseWriter, r *http.Request, libraryID, shelfID int) {
	shelf, err := h.bookshelfService.GetBookshelfByID(r.Context(), libraryID, shelfID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	resp := dto.BookshelfResponse{
		ID:         shelf.ID,
		LibraryID:  shelf.LibraryID,
		Code:       shelf.Code,
		Capacity:   shelf.Capacity,
		EmptySpace: shelf.EmptySpace,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /libraries/{library_id}/bookshelves/{shelf_id}/books
func (h *BookshelfHandler) GetBooksByShelfID(w http.ResponseWriter, r *http.Request, libraryID, shelfID int) {
	books, err := h.bookshelfService.GetBooksByShelfID(r.Context(), libraryID, shelfID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Handle case where no books exist on the shelf (return empty array [] instead of null)
	if books == nil {
		books = []dto.BookWithShelfStockResponse{}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(books)
}

// GET /libraries/{library_id}/bookshelves/{shelf_id}/books/{books_id}
func (h *BookshelfHandler) GetBookByShelfID(w http.ResponseWriter, r *http.Request, libraryID, shelfID, bookID int) {

}

// DELETE /libraries/{library_id}/bookshelves/{shelf_id}
func (h *BookshelfHandler) DeleteBookshelf(w http.ResponseWriter, r *http.Request, libraryID, shelfID int) {
	if err := h.bookshelfService.DeleteBookshelf(r.Context(), libraryID, shelfID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
