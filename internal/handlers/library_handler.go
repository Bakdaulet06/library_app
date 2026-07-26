package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"library/internal/dto"
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
	libraries, err := h.libraryService.ListLibraries(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

	books, err := h.libraryService.GetLibraryBooks(r.Context(), libraryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
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

func (h *LibraryHandler) GetLibraryBooksByGenre(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	genreIdStr := r.PathValue("genre_id")
	libraryID, err := strconv.Atoi(idStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}
	genreID, err := strconv.Atoi(genreIdStr)
	if err != nil || genreID <= 0 {
		http.Error(w, `{"error":"invalid genre ID"}`, http.StatusBadRequest)
		return
	}
	books, err := h.libraryService.GetLibraryBooksByGenre(r.Context(), libraryID, genreID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(books)
}

// Transactional Action endpoints
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
	var req dto.BorrowBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "malformed json request payload structure"})
		return
	}

	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := h.libraryService.BorrowBook(r.Context(), req, libraryID, bookID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "book processing successfully leased and logged"})
}
