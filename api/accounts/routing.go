package accounts

import (
	"github.com/keratin/authn-server/api"
)

func Routes(app *api.App) []*api.HandledRoute {
	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)
	authentication := api.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	// POST   /accounts/import

	return []*api.HandledRoute{
		api.Post("/accounts").
			SecuredWith(refererSecurity).
			Handle(postAccount(app)),

		api.Get("/accounts/available").
			SecuredWith(refererSecurity).
			Handle(getAccountsAvailable(app)),

		api.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(patchAccountLock(app)),

		api.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(patchAccountUnlock(app)),

		api.Patch("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(patchAccountExpirePassword(app)),

		api.Delete("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(deleteAccount(app)),
	}
}
