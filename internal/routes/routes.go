package routes

import (
	"net/http"

	"library/internal/handlers"
	"library/internal/middleware"
	"library/internal/models"
)

// Dependencies bundles every handler and middleware the route registrars need.
// Build one of these in main.go and pass it to RegisterAll.
type Dependencies struct {
	BookHandler          *handlers.BookHandler
	MemberHandler        *handlers.MemberHandler
	BookInventoryHandler *handlers.BookInventoryHandler
	LibraryHandler       *handlers.LibraryHandler
	BookshelfHandler     *handlers.BookshelfHandler
	UserHandler          *handlers.UserHandler
	EmployeeHandler      *handlers.EmployeeHandler

	AuthMiddleware func(http.Handler) http.Handler
}

// RegisterAll wires every route group onto mux. Call this once from main.go.
func RegisterAll(mux *http.ServeMux, deps Dependencies) {
	registerAuthRoutes(mux, deps)
	registerLibraryRoutes(mux, deps)
	registerBookshelfRoutes(mux, deps)
	registerBookRoutes(mux, deps)
	registerInventoryRoutes(mux, deps)
	registerMemberRoutes(mux, deps)
	registerEmployeeRoutes(mux, deps)
	registerLoanRoutes(mux, deps)
}

// protected applies auth middleware, and optionally role-based middleware,
// to a handler. Shared by every file in this package.
func protected(deps Dependencies, handler http.HandlerFunc, roles ...models.Role) http.Handler {
	if len(roles) > 0 {
		return middleware.Chain(handler, deps.AuthMiddleware, middleware.RequireRoles(roles...))
	}
	return middleware.Chain(handler, deps.AuthMiddleware)
}
