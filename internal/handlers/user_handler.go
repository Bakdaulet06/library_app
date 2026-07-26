package handlers

import (
	"encoding/json"
	"net/http"

	"library/internal/dto"
	"library/internal/services"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// POST /api/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "email and password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(r.Context(), req)
	if err != nil {
		// Handle duplicate email or validation errors
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// POST /api/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "email and password are required"}`, http.StatusBadRequest)
		return
	}

	authRes, err := h.userService.Login(r.Context(), req)
	if err != nil {
		// Generic invalid credentials message for security
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authRes)
}

// func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
// 	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
// 	if !ok {
// 		http.Error(w, `{"error": "user ID not found in context"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	// Fetch profile for userID...
// }
