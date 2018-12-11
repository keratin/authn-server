package accounts

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{}

	if app.Config.EnableSignup {
		routes = append(routes,
			route.Post("/accounts").
				SecuredWith(originSecurity).
				Handle(postAccount(app)),
			route.Get("/accounts/available").
				SecuredWith(originSecurity).
				Handle(getAccountsAvailable(app)),
		)
	}

	return routes
}

func Routes(app *app.App) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := PublicRoutes(app)

	routes = append(routes,
		route.Post("/accounts/import").
			SecuredWith(authentication).
			Handle(postAccountsImport(app)),

		route.Get("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(getAccount(app)),

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
