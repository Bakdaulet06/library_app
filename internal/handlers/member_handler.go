package handlers

import (
	"encoding/json"
	"library/internal/dto"
	"library/internal/services"
	"net/http"
	"strconv"
)

type MemberHandler struct {
	memberService services.MemberService
}

func NewMemberHandler(s services.MemberService) *MemberHandler {
	return &MemberHandler{memberService: s}
}

func (h *MemberHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid member ID"}`, http.StatusBadRequest)
		return
	}
	member, err := h.memberService.GetMemberByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(member)
}

func (h *MemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.memberService.ListMembers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal runtime exception listing profiles"})
		return
	}
	json.NewEncoder(w).Encode(members)
}

func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid member ID"}`, http.StatusBadRequest)
		return
	}
	var req dto.RegisterRequest
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

	member, err := h.memberService.UpdateMember(r.Context(), id, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(member)
}

func (h *MemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid member ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.memberService.DeleteMember(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MemberHandler) GetMemberLoans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID <= 0 {
		http.Error(w, `{"error":"invalid member ID"}`, http.StatusBadRequest)
		return
	}
	loans, err := h.memberService.GetMemberLoans(r.Context(), memberID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(loans)
}
