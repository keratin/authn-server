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
				Handle(PostAccount(app)),
			route.Get("/accounts/available").
				SecuredWith(originSecurity).
				Handle(GetAccountsAvailable(app)),
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
			Handle(PostAccountsImport(app)),

		route.Get("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(GetAccount(app)),

		route.Patch("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(PatchAccount(app)),

		route.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(PatchAccountLock(app)),

		route.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(PatchAccountUnlock(app)),

		route.Patch("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(PatchAccountExpirePassword(app)),

		route.Delete("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(DeleteAccount(app)),
	)

	return routes
}
