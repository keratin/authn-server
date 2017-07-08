package accounts

import (
	"github.com/keratin/authn-server/api"
)

func Routes(app *api.App) []*api.HandledRoute {
	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)
	authentication := api.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	// POST   /accounts/import
	// GET    /accounts/available
	// DELETE /accounts/{account_id:[0-9]+}

	return []*api.HandledRoute{
		api.Post("/accounts").
			SecuredWith(refererSecurity).
			Handle(postAccount(app)),

		api.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(patchAccountLock(app)),

		api.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(patchAccountUnlock(app)),
	}
}
