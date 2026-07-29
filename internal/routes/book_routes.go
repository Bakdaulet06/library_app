package routes

import (
	"net/http"

	"library/internal/models"
)

func registerBookRoutes(mux *http.ServeMux, deps Dependencies) {
	// Authenticated User Routes (any logged-in user can view/list)
	mux.Handle("GET /books", protected(deps, deps.BookHandler.ListBooks))
	mux.Handle("GET /books/genres/{id}", protected(deps, deps.BookHandler.GetBooksByGenreID))
	mux.Handle("GET /books/{id}", protected(deps, deps.BookHandler.GetBook))

	// Admin-only routes
	mux.Handle("POST /books", protected(deps, deps.BookHandler.CreateBook, models.RoleAdmin))
	mux.Handle("PUT /books/{id}", protected(deps, deps.BookHandler.UpdateBook, models.RoleAdmin))
	mux.Handle("DELETE /books/{id}", protected(deps, deps.BookHandler.DeleteBook, models.RoleAdmin))
}
