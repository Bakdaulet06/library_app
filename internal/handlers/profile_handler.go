package handlers

import (
	"encoding/json"
	"net/http"

	"library/internal/dto"
	"library/internal/middleware"
	"library/internal/services"
)

type ProfileHandler struct {
	profileService services.ProfileService // Adjust service interface name if different
}

func NewProfileHandler(profileService services.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// Request payload types for deposit/withdraw operations
type MoneyRequest struct {
	Amount float64 `json:"amount"`
}

// ----------------------------------------------------
// Profile Endpoints
// ----------------------------------------------------

// GET /profile
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Retrieve user claims/ID set by your Auth middleware context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	profile, err := h.profileService.GetProfileByUserID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// PUT /profile
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	var updateData dto.UpdateClientProfile // Or a specific update struct like UpdateProfileDTO
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	updatedProfile, err := h.profileService.UpdateProfile(r.Context(), userID, updateData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProfile)
}

// DELETE /profile
func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	if err := h.profileService.DeleteProfile(r.Context(), userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile successfully deleted"})
}

func (h *ProfileHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	if err := h.profileService.DeleteCard(r.Context(), userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile successfully deleted"})
}

// ----------------------------------------------------
// Card Endpoints
// ----------------------------------------------------

// GET /profile/card
func (h *ProfileHandler) GetCardDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	card, err := h.profileService.GetCardByUserID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(card)
}

// POST /profile/card
func (h *ProfileHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	// No request body parsing needed since amount defaults to 0
	card, err := h.profileService.CreateCard(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

// POST /profile/card/withdraw
func (h *ProfileHandler) WithdrawMoney(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	var req MoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, `{"error":"invalid amount for withdrawal"}`, http.StatusBadRequest)
		return
	}

	updatedCard, err := h.profileService.WithdrawMoney(r.Context(), userID, req.Amount)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Withdrawal successful",
		"card":    updatedCard,
	})
}

// POST /profile/card/deposit
func (h *ProfileHandler) DepositMoney(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized access"}`, http.StatusUnauthorized)
		return
	}

	userID := user.ID

	var req MoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, `{"error":"invalid amount for deposit"}`, http.StatusBadRequest)
		return
	}

	updatedCard, err := h.profileService.DepositMoney(r.Context(), userID, req.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Deposit successful",
		"card":    updatedCard,
	})
}
