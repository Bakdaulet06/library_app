package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"library/internal/dto"
	"library/internal/services"
)

type BookInventoryHandler struct {
	inventoryService services.BookInventoryService
	bookshelfService services.BookshelfService
}

func NewBookInventoryHandler(inventoryService services.BookInventoryService, bookshelfService services.BookshelfService) *BookInventoryHandler {
	return &BookInventoryHandler{
		inventoryService: inventoryService,
		bookshelfService: bookshelfService,
	}
}

func (h *BookInventoryHandler) AddInventory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inventory, err := h.inventoryService.AddBookInventory(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inventory)
}

func (h *BookInventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	inventory, err := h.inventoryService.ListBookInventory(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

func (h *BookInventoryHandler) GetAvailableCopies(w http.ResponseWriter, r *http.Request) {
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
	copies, err := h.inventoryService.GetAvailableCopies(r.Context(), bookID, libraryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"library_id":       libraryID,
		"book_id":          bookID,
		"available_copies": *copies,
	})
}

func (h *BookInventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
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
	if err := h.inventoryService.DeleteBookInventory(r.Context(), libraryID, bookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
