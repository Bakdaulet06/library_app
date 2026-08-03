package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"library/internal/dto"
	"library/internal/middleware"
	"library/internal/models"
	"library/internal/params"
	"library/internal/services"
)

type LibraryHandler struct {
	libraryService   services.LibraryService
	inventoryService services.BookInventoryService
}

func NewLibraryHandler(libraryService services.LibraryService, inventoryService services.BookInventoryService) *LibraryHandler {
	return &LibraryHandler{libraryService: libraryService, inventoryService: inventoryService}
}

func (h *LibraryHandler) RegisterLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req dto.CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	library, err := h.libraryService.RegisterLibrary(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(library)
}

func (h *LibraryHandler) ListLibraries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query params (limit, offset, sort_by, order, q)
	p := params.FromRequest(r)

	libraries, err := h.libraryService.ListLibraries(r.Context(), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(libraries)
}

func (h *LibraryHandler) GetLibraryByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	library, err := h.libraryService.GetLibraryByID(r.Context(), libraryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(library)
}

func (h *LibraryHandler) UpdateLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}

	var req dto.CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updated, err := h.libraryService.UpdateLibrary(r.Context(), libraryID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *LibraryHandler) DeleteLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.libraryService.DeleteLibrary(r.Context(), libraryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) DeleteBookFromLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	bookIdStr := r.PathValue("book_id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	bookID, err := strconv.Atoi(bookIdStr)
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"invalid book ID"}`, http.StatusBadRequest)
		return
	}
	// Calls your inventory service to delete the stock entry from book_inventory
	if err := h.inventoryService.DeleteBookInventory(r.Context(), libraryID, bookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) GetLibraryBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}

	// Parse both GenreID and standard Pagination params
	p := params.BookParamsFromRequest(r)

	books, err := h.libraryService.GetLibraryBooks(r.Context(), libraryID, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if books == nil {
		books = []models.LibraryBook{}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(books)
}

func (h *LibraryHandler) GetLibraryLoans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	loans, err := h.libraryService.GetLibraryLoans(r.Context(), libraryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loans)
}

// Transactional Action endpoints
// for request we send for much time user want to loan the book, starting from 7 days till 21 days,
type borrowBookRequest struct {
	BorrowDays int `json:"borrow_days"`
}

func (h *LibraryHandler) BorrowBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	bookIdStr := r.PathValue("book_id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	bookID, err := strconv.Atoi(bookIdStr)
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"invalid book ID"}`, http.StatusBadRequest)
		return
	}

	var req borrowBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: missing or invalid user identity"})
		return
	}

	if err := h.libraryService.BorrowBook(r.Context(), user.ID, libraryID, bookID, req.BorrowDays); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "book processing successfully leased and logged"})
}

func (h *LibraryHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: missing or invalid user identity"})
		return
	}
	var req dto.ReturnBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid return book request payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := h.libraryService.ReturnBook(r.Context(), libraryID, user.ID, req.BookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "book processing successfully leased and logged"})
}

func (h *LibraryHandler) ListReturnedBooks(w http.ResponseWriter, r *http.Request) {
	libraryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid library id", http.StatusBadRequest)
		return
	}

	returnedBooks, err := h.libraryService.ListReturnedBooks(r.Context(), libraryID)
	if err != nil {
		http.Error(w, "failed to fetch returned books: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(returnedBooks); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *LibraryHandler) AssignShelf(w http.ResponseWriter, r *http.Request) {
	libraryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid library id", http.StatusBadRequest)
		return
	}

	bookID, err := strconv.Atoi(r.PathValue("book_id"))
	if err != nil {
		http.Error(w, "invalid book id", http.StatusBadRequest)
		return
	}

	shelf, err := h.libraryService.AssignShelf(r.Context(), libraryID, bookID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "no pending returned record"):
			http.Error(w, err.Error(), http.StatusNotFound)
		case strings.Contains(err.Error(), "no available shelf space"):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "failed to assign shelf: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(shelf); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
