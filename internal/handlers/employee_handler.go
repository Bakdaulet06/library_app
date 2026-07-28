package handlers

import (
	"encoding/json"
	"library/internal/dto"
	"library/internal/services"
	"net/http"
	"strconv"
)

type EmployeeHandler struct {
	employeeService services.EmployeeService
}

func NewEmployeeHandler(s services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employeeService: s}
}

func (h *EmployeeHandler) RegisterEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req dto.CreateEmployeeRequest
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

	emp, err := h.employeeService.RegisterEmployee(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(emp)
}

func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID <= 0 {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}
	emp, err := h.employeeService.GetEmployeeByMemberID(r.Context(), memberID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(emp)
}

func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	employees, err := h.employeeService.ListEmployees(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal runtime exception listing profiles"})
		return
	}
	json.NewEncoder(w).Encode(employees)
}

func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID <= 0 {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}
	var req dto.UpdateEmployeeRequest
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

	emp, err := h.employeeService.UpdateEmployee(r.Context(), memberID, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(emp)
}

func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID <= 0 {
		http.Error(w, `{"error":"invalid employee ID"}`, http.StatusBadRequest)
		return
	}
	if err := h.employeeService.DeleteEmployee(r.Context(), memberID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
