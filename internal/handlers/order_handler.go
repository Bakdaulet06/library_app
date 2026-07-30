package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"library/internal/dto"
	"library/internal/services"

	// TODO: replace with wherever your auth middleware actually lives -
	// this is a placeholder import name, adjust to match your project.
	"library/internal/middleware"
)

type OrderHandler struct {
	orderService   services.OrderService
	profileService services.ProfileService
}

func NewOrderHandler(orderService services.OrderService, profileService services.ProfileService) *OrderHandler {
	return &OrderHandler{orderService: orderService, profileService: profileService}
}

// BuyBook handles POST /libraries/{id}/books/{book_id}/buy
func (h *OrderHandler) BuyBook(w http.ResponseWriter, r *http.Request) {
	libraryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || libraryID <= 0 {
		http.Error(w, "invalid library id", http.StatusBadRequest)
		return
	}

	bookID, err := strconv.Atoi(r.PathValue("book_id"))
	if err != nil || bookID <= 0 {
		http.Error(w, "invalid book id", http.StatusBadRequest)
		return
	}
	member, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	memberID := member.ID

	var req dto.BuyBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body, should include only quantity", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: this assumes middleware.Authenticate (or similar) stashes the
	// authenticated member's ID in the request context, and exposes a
	// helper to pull it back out. Rename this call to match whatever your
	// actual auth middleware provides - e.g. it might instead be
	// middleware.UserFromContext(r.Context()).ID or similar.

	card, err := h.profileService.GetCardByUserID(r.Context(), memberID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	buyBookRequestFull := dto.BuyBookRequestFull{
		MemberID:  memberID,
		LibraryID: libraryID,
		BookID:    bookID,
		CardID:    card.ID,
		Quantity:  req.Quantity,
	}

	result, err := h.orderService.BuyBooks(r.Context(), buyBookRequestFull)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientStock):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "failed to complete order", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.BuyBookResponse{
		OrderID:     result.OrderID,
		LibraryID:   result.LibraryID,
		BookID:      result.BookID,
		Quantity:    result.Quantity,
		UnitPrice:   result.UnitPrice,
		TotalAmount: result.TotalAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) ListOrdersOfSpecificClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Extract authenticated user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}
	userID := user.ID

	// 2. Fetch user's orders via service
	orders, err := h.orderService.GetOrdersByMemberID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

// GET /libraries/{library_id}/orders
// Retrieves all orders placed at a specific library (for Staff/Admin)
func (h *OrderHandler) ListOrdersByLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Extract library_id from URL path or query params
	// If using Go 1.22 standard mux path value: libraryIDStr := r.PathValue("library_id")
	// Or via URL query parameter: r.URL.Query().Get("library_id")
	libraryIDStr := r.PathValue("id")
	if libraryIDStr == "" {
		libraryIDStr = r.URL.Query().Get("id")
	}

	libraryID, err := strconv.Atoi(libraryIDStr)
	if err != nil || libraryID <= 0 {
		http.Error(w, `{"error":"invalid or missing library id"}`, http.StatusBadRequest)
		return
	}

	// 2. Fetch library orders via service
	orders, err := h.orderService.GetOrdersByLibraryID(r.Context(), libraryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

// GET /orders
// Retrieves all orders across the system (Admin only)
func (h *OrderHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orders, err := h.orderService.GetAllOrders(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
