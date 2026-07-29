package routes

import (
	"net/http"

	"library/internal/models"
)

func registerEmployeeRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /employees", protected(deps, deps.EmployeeHandler.ListEmployees, models.RoleAdmin))
	mux.Handle("GET /employees/{member_id}", protected(deps, deps.EmployeeHandler.GetEmployee, models.RoleAdmin))
	mux.Handle("POST /employees", protected(deps, deps.EmployeeHandler.RegisterEmployee, models.RoleAdmin))
	mux.Handle("PUT /employees/{member_id}", protected(deps, deps.EmployeeHandler.UpdateEmployee, models.RoleAdmin))
	mux.Handle("DELETE /employees/{member_id}", protected(deps, deps.EmployeeHandler.DeleteEmployee, models.RoleAdmin))
}
