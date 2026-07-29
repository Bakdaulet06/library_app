package routes

import (
	"net/http"

	"library/internal/models"
)

func registerBookshelfRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /libraries/{id}/bookshelves", protected(deps, deps.BookshelfHandler.GetBookshelvesByLibraryID))
	mux.Handle("POST /libraries/{id}/bookshelves", protected(deps, deps.BookshelfHandler.CreateBookshelf, models.RoleAdmin))
	mux.Handle("GET /libraries/{id}/bookshelves/{shelf_id}", protected(deps, deps.BookshelfHandler.GetBookshelfByID))
	mux.Handle("DELETE /libraries/{id}/bookshelves/{shelf_id}", protected(deps, deps.BookshelfHandler.DeleteBookshelf, models.RoleAdmin))
	mux.Handle("GET /libraries/{id}/bookshelves/{shelf_id}/books", protected(deps, deps.BookshelfHandler.GetBooksByShelfID))
}
