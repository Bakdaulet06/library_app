package routes

import (
	"net/http"

	"library/internal/models"
)

func registerLoanRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("/loans", protected(deps, deps.BookHandler.HandleListLoans, models.RoleAdmin))
}
