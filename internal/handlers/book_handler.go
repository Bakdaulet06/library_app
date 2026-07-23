package handlers

import (
	"encoding/json"
	"library/internal/dto"
	"library/internal/services"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type BookHandler struct {
	service services.BookService
}

func NewBookHandler(s services.BookService) *BookHandler {
	return &BookHandler{service: s}
}

// ServeHTTP handles all routed routing variations for /books and /books/ paths
func (h *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/books")
	path = strings.Trim(path, "/")

	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		// 1. Root: /books
		if path == "" {
			h.ListBooks(w, r)
			return
		}

		// 2. Sub-resource: /books/genres/{id}
		if len(parts) == 2 && parts[0] == "genres" {
			genreID, err := strconv.Atoi(parts[1])
			if err != nil {
				http.Error(w, `{"error":"invalid genre ID format"}`, http.StatusBadRequest)
				return
			}
			h.GetBooksByGenreID(w, r, genreID)
			return
		}

		// 3. Single resource: /books/{id}
		if id, err := strconv.Atoi(path); err == nil {
			h.GetBook(w, r, id)
			return
		}

	case http.MethodPost:
		if path == "" {
			h.CreateBook(w, r)
			return
		}

	case http.MethodPut:
		if id, err := strconv.Atoi(path); err == nil {
			h.UpdateBook(w, r, id)
			return
		}

	case http.MethodDelete:
		if id, err := strconv.Atoi(path); err == nil {
			h.DeleteBook(w, r, id)
			return
		}
	}

	http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookRequest
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

	book, err := h.service.CreateBook(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request, id int) {
	book, err := h.service.GetBookByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	// Optional Query Lookup Parameter filter: /books?available=true
	availableFilter := r.URL.Query().Get("available")

	var books interface{}
	var err error

	if availableFilter == "true" {
		books, err = h.service.ListAvailableBooks(r.Context())
	} else {
		books, err = h.service.ListAllBooks(r.Context())
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal runtime exception listing books"})
		return
	}
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request, id int) {
	var req dto.CreateBookRequest
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

	book, err := h.service.UpdateBook(r.Context(), id, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.service.DeleteBook(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Transactional Action endpoints

func (h *BookHandler) HandleBorrow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed, use POST"}`, http.StatusMethodNotAllowed)
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

	if err := h.service.BorrowBook(r.Context(), req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "book processing successfully leased and logged"})
}

func (h *BookHandler) HandleReturn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed, use POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var req dto.ReturnBookRequest
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

	if err := h.service.ReturnBook(r.Context(), req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "asset safely returned and inventory incremented"})
}

// Append this function to the bottom of internal/handlers/book_handler.go
func (h *BookHandler) HandleListLoans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed, use GET"}`, http.StatusMethodNotAllowed)
		return
	}

	loans, err := h.service.ListAllLoans(r.Context())
	if err != nil {
		log.Printf("ERROR inside ListAllLoans: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(), // <--- Temporarily send real error to Postman
		})
		return
	}

	json.NewEncoder(w).Encode(loans)
}

func (h *BookHandler) GetBooksByGenreID(w http.ResponseWriter, r *http.Request, genreID int) {
	books, err := h.service.GetBooksByGenreID(r.Context(), genreID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(books)
}
