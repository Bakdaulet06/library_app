package handlers

import (
	"encoding/json"
	"net/http"

	"library/internal/dto"
	"library/internal/services"
)

type MemberHandler struct {
	service services.MemberService
}

func NewMemberHandler(s services.MemberService) *MemberHandler {
	return &MemberHandler{service: s}
}

// ServeHTTP acts as the primary multiplexer for member-related endpoints
func (h *MemberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/members" {
		if r.Method == http.MethodPost {
			h.registerMember(w, r)
			return
		}
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	respondWithError(w, http.StatusNotFound, "resource path not found")
}

func (h *MemberHandler) registerMember(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "malformed json payload")
		return
	}

	if err := req.Validate(); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	member, err := h.service.RegisterMember(r.Context(), req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}
