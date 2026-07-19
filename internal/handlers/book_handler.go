package handlers

import (
	"encoding/json"
	"net/http"

	"library/internal/dto"
	"library/internal/models"
	"library/internal/services"
)

type BookHandler struct {
	service services.BookService
}

func NewBookHandler(s services.BookService) *BookHandler {
	return &BookHandler{service: s}
}

// ServeHTTP acts as the primary multiplexer for book-related endpoints
func (h *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/books":
		if r.Method == http.MethodPost {
			h.createBook(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.listAvailable(w, r)
			return
		}
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "/books/borrow":
		if r.Method == http.MethodPost {
			h.borrowBook(w, r)
			return
		}
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")

	case "/books/return":
		if r.Method == http.MethodPost {
			h.returnBook(w, r)
			return
		}
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")

	default:
		respondWithError(w, http.StatusNotFound, "resource path not found")
	}
}

func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "malformed json payload")
		return
	}

	if err := req.Validate(); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	book, err := h.service.CreateBook(r.Context(), req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) listAvailable(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.ListAvailableBooks(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Graceful fallback response: return an empty JSON array instead of a null value
	if books == nil {
		books = make([]models.Book, 0)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) borrowBook(w http.ResponseWriter, r *http.Request) {
	var req dto.BorrowBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "malformed json payload")
		return
	}

	if err := req.Validate(); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.service.BorrowBook(r.Context(), req); err != nil {
		// Differentiating client domain exceptions from true server crashes
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "book successfully checked out"})
}

func (h *BookHandler) returnBook(w http.ResponseWriter, r *http.Request) {
	var req dto.ReturnBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "malformed json payload")
		return
	}

	if err := req.Validate(); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.service.ReturnBook(r.Context(), req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "book successfully returned to inventory"})
}

// Global internal JSON error writer helper
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
