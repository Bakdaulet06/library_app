package routes

import (
	"net/http"

	"library/internal/models"
)

func registerMemberRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /clients", protected(deps, deps.MemberHandler.ListMembers, models.RoleAdmin))
	mux.Handle("POST /clients", protected(deps, deps.UserHandler.RegisterClient, models.RoleAdmin))
	mux.Handle("GET /clients/{id}", protected(deps, deps.MemberHandler.GetMember, models.RoleAdmin))
	mux.Handle("GET /clients/{id}/loans", protected(deps, deps.MemberHandler.GetMemberLoans, models.RoleAdmin))
	mux.Handle("PUT /clients/{id}", protected(deps, deps.MemberHandler.UpdateMember, models.RoleAdmin))
	mux.Handle("DELETE /clients/{id}", protected(deps, deps.MemberHandler.DeleteMember, models.RoleAdmin))
}
