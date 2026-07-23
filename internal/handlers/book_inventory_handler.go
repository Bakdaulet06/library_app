package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"library/internal/dto"
	"library/internal/services"
)

type BookInventoryHandler struct {
	inventoryService services.BookInventoryService
}

func NewBookInventoryHandler(inventoryService services.BookInventoryService) *BookInventoryHandler {
	return &BookInventoryHandler{inventoryService: inventoryService}
}

// ServeHTTP handles /inventory and /inventory/{libraryId}/{bookId}
func (h *BookInventoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/inventory")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListInventory(w, r)
		case http.MethodPost:
			h.CreateInventory(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Parsing composite key route: /inventory/{libraryId}/{bookId}
	parts := strings.Split(path, "/")
	if len(parts) == 2 {
		libraryID, err1 := strconv.Atoi(parts[0])
		bookID, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			http.Error(w, "Invalid library or book ID format", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.GetAvailableCopies(w, r, libraryID, bookID)
		case http.MethodPut:
			h.UpdateInventory(w, r, libraryID, bookID)
		case http.MethodDelete:
			h.DeleteInventory(w, r, libraryID, bookID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.Error(w, "Route not found", http.StatusNotFound)
}

func (h *BookInventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inventory, err := h.inventoryService.CreateOrUpdateBookInventory(r.Context(), req)
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

func (h *BookInventoryHandler) GetAvailableCopies(w http.ResponseWriter, r *http.Request, libraryID, bookID int) {
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

func (h *BookInventoryHandler) UpdateInventory(w http.ResponseWriter, r *http.Request, libraryID, bookID int) {
	var req dto.CreateBookInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updated, err := h.inventoryService.CreateOrUpdateBookInventory(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *BookInventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request, libraryID, bookID int) {
	if err := h.inventoryService.DeleteBookInventory(r.Context(), libraryID, bookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
