package routes

import (
	"net/http"

	"library/internal/models"
)

func registerGenreRoutes(mux *http.ServeMux, deps Dependencies) {
	// Public or standard authenticated access (anyone logged in can view genres)
	mux.Handle("GET /genres", protected(deps, deps.GenreHandler.GetAll, models.RoleAdmin))
	mux.Handle("GET /genres/{id}", protected(deps, deps.GenreHandler.GetByID, models.RoleAdmin))

	// Admin-only mutation routes
	mux.Handle("POST /genres", protected(deps, deps.GenreHandler.Create, models.RoleAdmin))
	mux.Handle("PUT /genres/{id}", protected(deps, deps.GenreHandler.Update, models.RoleAdmin))
	mux.Handle("DELETE /genres/{id}", protected(deps, deps.GenreHandler.Delete, models.RoleAdmin))
}
