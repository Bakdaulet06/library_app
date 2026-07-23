package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"library/internal/dto"
	"library/internal/services"
)

type LibraryHandler struct {
	libraryService   services.LibraryService
	inventoryService services.BookInventoryService
}

func NewLibraryHandler(libraryService services.LibraryService) *LibraryHandler {
	return &LibraryHandler{libraryService: libraryService}
}

// ServeHTTP handles /libraries and /libraries/{id} or /libraries/{id}/books, /libraries/{id}/loans
func (h *LibraryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/libraries")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListLibraries(w, r)
		case http.MethodPost:
			h.RegisterLibrary(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	parts := strings.Split(path, "/")
	libraryID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid library ID"}`, http.StatusBadRequest)
		return
	}

	// 1. Handle 4-part path: /libraries/{libraryId}/books/genres/{genreId}
	if len(parts) == 4 && parts[1] == "books" && parts[2] == "genres" {
		genreID, err := strconv.Atoi(parts[3])
		if err != nil {
			http.Error(w, `{"error":"invalid genre ID"}`, http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			h.GetLibraryBooksByGenre(w, r, libraryID, genreID)
			return
		}

		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 2. Handle 3-part path: /libraries/{libraryId}/books/{bookId}
	if len(parts) == 3 && parts[1] == "books" {
		bookID, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, `{"error":"invalid book ID"}`, http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodDelete {
			h.DeleteBookFromLibrary(w, r, libraryID, bookID)
			return
		}

		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 3. Handle 2-part paths: /libraries/{id}/books or /libraries/{id}/loans
	if len(parts) == 2 {
		switch parts[1] {
		case "books":
			if r.Method == http.MethodGet {
				h.GetLibraryBooks(w, r, libraryID)
				return
			}
		case "loans":
			if r.Method == http.MethodGet {
				h.GetLibraryLoans(w, r, libraryID)
				return
			}
		}
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 4. Handle 1-part path: /libraries/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetLibraryByID(w, r, libraryID)
		case http.MethodPut:
			h.UpdateLibrary(w, r, libraryID)
		case http.MethodDelete:
			h.DeleteLibrary(w, r, libraryID)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
		return
	}

	http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
}

func (h *LibraryHandler) RegisterLibrary(w http.ResponseWriter, r *http.Request) {
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
	libraries, err := h.libraryService.ListLibraries(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(libraries)
}

func (h *LibraryHandler) GetLibraryByID(w http.ResponseWriter, r *http.Request, id int) {
	library, err := h.libraryService.GetLibraryByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(library)
}

func (h *LibraryHandler) UpdateLibrary(w http.ResponseWriter, r *http.Request, id int) {
	var req dto.CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updated, err := h.libraryService.UpdateLibrary(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *LibraryHandler) DeleteLibrary(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.libraryService.DeleteLibrary(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) DeleteBookFromLibrary(w http.ResponseWriter, r *http.Request, libraryID, bookID int) {
	// Calls your inventory service to delete the stock entry from book_inventory
	if err := h.inventoryService.DeleteBookInventory(r.Context(), libraryID, bookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) GetLibraryBooks(w http.ResponseWriter, r *http.Request, id int) {
	books, err := h.libraryService.GetLibraryBooks(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *LibraryHandler) GetLibraryLoans(w http.ResponseWriter, r *http.Request, id int) {
	loans, err := h.libraryService.GetLibraryLoans(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loans)
}

func (h *LibraryHandler) GetLibraryBooksByGenre(w http.ResponseWriter, r *http.Request, libraryID, genreID int) {
	books, err := h.libraryService.GetLibraryBooksByGenre(r.Context(), libraryID, genreID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(books)
}
