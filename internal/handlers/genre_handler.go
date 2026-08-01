package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"library/internal/dto"
	"library/internal/repositories"
	"library/internal/services"
)

type GenreHandler struct {
	service services.GenreService
}

func NewGenreHandler(service services.GenreService) *GenreHandler {
	return &GenreHandler{service: service}
}

func (h *GenreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrUpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	genre, err := h.service.CreateGenre(r.Context(), req)
	if err != nil {
		if errors.Is(err, repositories.ErrGenreExists) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "failed to create genre"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
}

func (h *GenreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid genre ID"}`, http.StatusBadRequest)
		return
	}

	genre, err := h.service.GetGenreByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrGenreNotFound) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
}

func (h *GenreHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	genres, err := h.service.GetAllGenres(r.Context())
	if err != nil {
		http.Error(w, `{"error": "failed to fetch genres"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid genre ID"}`, http.StatusBadRequest)
		return
	}

	var req dto.CreateOrUpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	genre, err := h.service.UpdateGenre(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, repositories.ErrGenreNotFound) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "failed to update genre"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
}

func (h *GenreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid genre ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteGenre(r.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrGenreNotFound) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "failed to delete genre"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
