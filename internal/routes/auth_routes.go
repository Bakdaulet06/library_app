package routes

import (
	"net/http"

	"library/internal/models"
)

func registerAuthRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.HandleFunc("POST /register", deps.UserHandler.Register)
	mux.HandleFunc("POST /login", deps.UserHandler.Login)
	mux.Handle("POST /register_client", protected(deps, deps.UserHandler.Register, models.RoleAdmin, models.RoleEmployee))
}
