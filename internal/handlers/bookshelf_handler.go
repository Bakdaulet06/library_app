package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// POST /libraries/{library_id}/bookshelves
func (h *BookshelfHandler) CreateBookshelf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
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
func (h *BookshelfHandler) GetBookshelvesByLibraryID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
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
func (h *BookshelfHandler) GetBookshelfByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	shelfIDStr := r.PathValue("book_id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	shelfID, err := strconv.Atoi(shelfIDStr)
	if err != nil || shelfID <= 0 {
		http.Error(w, `{"error":"invalid shelf ID"}`, http.StatusBadRequest)
		return
	}
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
func (h *BookshelfHandler) GetBooksByShelfID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	shelfIDStr := r.PathValue("shelf_id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	shelfID, err := strconv.Atoi(shelfIDStr)
	if err != nil || shelfID <= 0 {
		http.Error(w, `{"error":"invalid shelf ID"}`, http.StatusBadRequest)
		return
	}

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

// DELETE /libraries/{library_id}/bookshelves/{shelf_id}
func (h *BookshelfHandler) DeleteBookshelf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	shelfIDStr := r.PathValue("shelf_id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	shelfID, err := strconv.Atoi(shelfIDStr)
	if err != nil || shelfID <= 0 {
		http.Error(w, `{"error":"invalid shelf ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.bookshelfService.DeleteBookshelf(r.Context(), libraryID, shelfID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
