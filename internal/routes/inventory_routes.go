package routes

import (
	"net/http"

	"library/internal/models"
)

func registerInventoryRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /inventory", protected(deps, deps.BookInventoryHandler.ListInventory, models.RoleAdmin))
	mux.Handle("POST /inventory", protected(deps, deps.BookInventoryHandler.AddInventory, models.RoleAdmin))
	mux.Handle("GET /inventory/{id}/{book_id}", protected(deps, deps.BookInventoryHandler.GetAvailableCopies, models.RoleAdmin))
	mux.Handle("DELETE /inventory/{id}/{book_id}", protected(deps, deps.BookInventoryHandler.DeleteInventory, models.RoleAdmin))
}
