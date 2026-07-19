package handlers

import (
	"encoding/json"
	"library/internal/dto"
	"library/internal/services"
	"net/http"
	"strconv"
	"strings"
)

type MemberHandler struct {
	service services.MemberService
}

func NewMemberHandler(s services.MemberService) *MemberHandler {
	return &MemberHandler{service: s}
}

func (h *MemberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/members")
	path = strings.Trim(path, "/")

	switch r.Method {
	case http.MethodPost:
		if path == "" {
			h.RegisterMember(w, r)
			return
		}
	case http.MethodGet:
		if path == "" {
			h.ListMembers(w, r)
			return
		}
		if id, err := strconv.ParseInt(path, 10, 64); err == nil {
			h.GetMember(w, r, id)
			return
		}
	case http.MethodPut:
		if id, err := strconv.ParseInt(path, 10, 64); err == nil {
			h.UpdateMember(w, r, id)
			return
		}
	case http.MethodDelete:
		if id, err := strconv.ParseInt(path, 10, 64); err == nil {
			h.DeleteMember(w, r, id)
			return
		}
	}

	http.Error(w, `{"error":"endpoint route or method pattern not found"}`, http.StatusNotFound)
}

func (h *MemberHandler) RegisterMember(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMemberRequest
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

	member, err := h.service.RegisterMember(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}

func (h *MemberHandler) GetMember(w http.ResponseWriter, r *http.Request, id int64) {
	member, err := h.service.GetMemberByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(member)
}

func (h *MemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.service.ListMembers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal runtime exception listing profiles"})
		return
	}
	json.NewEncoder(w).Encode(members)
}

func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request, id int64) {
	var req dto.CreateMemberRequest
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

	member, err := h.service.UpdateMember(r.Context(), id, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(member)
}

func (h *MemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.DeleteMember(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
