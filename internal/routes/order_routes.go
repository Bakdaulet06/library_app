package routes

import (
	"net/http"

	"library/internal/models"
)

func registerOrderRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("/orders", protected(deps, deps.OrderHandler.ListAllOrders, models.RoleAdmin))
}
