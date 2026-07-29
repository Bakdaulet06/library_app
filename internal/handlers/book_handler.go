package handlers

import (
	"encoding/json"
	"library/internal/dto"
	"library/internal/services"
	"log"
	"net/http"
	"strconv"
)

type BookHandler struct {
	service services.BookService
}

func NewBookHandler(s services.BookService) *BookHandler {
	return &BookHandler{service: s}
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

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bookID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"invalid book ID"}`, http.StatusBadRequest)
		return
	}
	book, err := h.service.GetBookByID(r.Context(), bookID)
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

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bookID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"invalid book ID"}`, http.StatusBadRequest)
		return
	}

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

	book, err := h.service.UpdateBook(r.Context(), bookID, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bookID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || bookID <= 0 {
		http.Error(w, `{"error":"invalid book ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteBook(r.Context(), bookID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Book successfully deleted"})
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

func (h *BookHandler) GetBooksByGenreID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	genreID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || genreID <= 0 {
		http.Error(w, `{"error":"invalid genre ID"}`, http.StatusBadRequest)
		return
	}
	books, err := h.service.GetBooksByGenreID(r.Context(), genreID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(books)
}
