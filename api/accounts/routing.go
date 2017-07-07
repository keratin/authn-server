package accounts

import (
	"github.com/keratin/authn-server/api"
)

func Routes(app *api.App) []*api.HandledRoute {
	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)

	// POST   /accounts/import
	// GET    /accounts/available
	// DELETE /accounts/{account_id:[0-9]+}
	// PATCH  /accounts/{account_id:[0-9]+}/lock
	// PATCH  /accounts/{account_id:[0-9]+}/unlock

	return []*api.HandledRoute{
		api.Post("/accounts").
			SecuredWith(refererSecurity).
			Handle(postAccount(app)),
	}
}
