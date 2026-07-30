package routes

import (
	"net/http"

	"library/internal/models"
)

func registerProfileRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /profile", protected(deps, deps.ProfileHandler.GetProfile, models.RoleClient))
	mux.Handle("PUT /profile", protected(deps, deps.ProfileHandler.UpdateProfile, models.RoleClient))
	mux.Handle("DELETE /profile", protected(deps, deps.ProfileHandler.DeleteProfile, models.RoleClient))

	mux.Handle("GET /profile/card", protected(deps, deps.ProfileHandler.GetCardDetails, models.RoleClient))
	mux.Handle("POST /profile/card", protected(deps, deps.ProfileHandler.CreateCard, models.RoleClient))
	mux.Handle("DELETE /profile/card", protected(deps, deps.ProfileHandler.DeleteCard, models.RoleClient))
	mux.Handle("POST /profile/card/withdraw", protected(deps, deps.ProfileHandler.WithdrawMoney, models.RoleClient))
	mux.Handle("POST /profile/card/deposit", protected(deps, deps.ProfileHandler.DepositMoney, models.RoleClient))

	mux.Handle("GET /profile/orders", protected(deps, deps.OrderHandler.ListOrdersOfSpecificClient, models.RoleClient))
}
