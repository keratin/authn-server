package accounts

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func Routes(app *api.App) []*route.HandledRoute {
	refererSecurity := route.RefererSecurity(app.Config.ApplicationDomains)
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := []*route.HandledRoute{}

	if app.Config.EnableSignup {
		routes = append(routes,
			route.Post("/accounts").
				SecuredWith(refererSecurity).
				Handle(postAccount(app)),
			route.Get("/accounts/available").
				SecuredWith(refererSecurity).
				Handle(getAccountsAvailable(app)),
		)
	}

	routes = append(routes,
		route.Post("/accounts/import").
			SecuredWith(authentication).
			Handle(postAccountsImport(app)),

		route.Patch("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(patchAccount(app)),

		route.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(patchAccountLock(app)),

		route.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(patchAccountUnlock(app)),

		route.Patch("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(patchAccountExpirePassword(app)),

		route.Delete("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(deleteAccount(app)),
	)

	return routes
}
