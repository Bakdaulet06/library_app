package routes

import (
	"net/http"

	"library/internal/models"
)

func registerLibraryRoutes(mux *http.ServeMux, deps Dependencies) {
	// ----------------------------------------------------
	// Libraries
	// ----------------------------------------------------
	mux.Handle("GET /libraries", protected(deps, deps.LibraryHandler.ListLibraries))
	mux.Handle("POST /libraries", protected(deps, deps.LibraryHandler.RegisterLibrary, models.RoleAdmin))
	mux.Handle("GET /libraries/{id}", protected(deps, deps.LibraryHandler.GetLibraryByID))
	mux.Handle("PUT /libraries/{id}", protected(deps, deps.LibraryHandler.UpdateLibrary, models.RoleAdmin))
	mux.Handle("DELETE /libraries/{id}", protected(deps, deps.LibraryHandler.DeleteLibrary, models.RoleAdmin))

	// ----------------------------------------------------
	// Library Books & Loans
	// ----------------------------------------------------
	mux.Handle("GET /libraries/{id}/books", protected(deps, deps.LibraryHandler.GetLibraryBooks))
	mux.Handle("GET /libraries/{id}/loans", protected(deps, deps.LibraryHandler.GetLibraryLoans, models.RoleAdmin, models.RoleEmployee))
	mux.Handle("DELETE /libraries/{id}/books/{book_id}", protected(deps, deps.LibraryHandler.DeleteBookFromLibrary, models.RoleAdmin))
	mux.Handle("GET /libraries/{id}/books/genres/{genre_id}", protected(deps, deps.LibraryHandler.GetLibraryBooksByGenre))
	mux.Handle("POST /libraries/{id}/books/{book_id}/borrow", protected(deps, deps.LibraryHandler.BorrowBook, models.RoleClient))
	mux.Handle("POST /libraries/{id}/books/{book_id}/buy", protected(deps, deps.OrderHandler.BuyBook, models.RoleClient))
	mux.Handle("POST /libraries/{id}/return", protected(deps, deps.LibraryHandler.ReturnBook, models.RoleClient))
	mux.Handle("GET /libraries/{id}/returned_books", protected(deps, deps.LibraryHandler.ListReturnedBooks, models.RoleEmployee))
	mux.Handle("POST /libraries/{id}/returned_books/{book_id}/assign_shelf", protected(deps, deps.LibraryHandler.AssignShelf, models.RoleEmployee))

	mux.Handle("GET /libraries/{id}/orders", protected(deps, deps.OrderHandler.ListOrdersByLibrary, models.RoleEmployee, models.RoleAdmin))
}
