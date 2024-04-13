package server

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PrivateRoutes(app *app.App) []*route.HandledRoute {
	var routes []*route.HandledRoute
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes = append(routes,
		route.Get("/").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetRoot(app)),

		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetJWKs(app)),

		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetConfiguration(app)),

		route.Get("/metrics").
			SecuredWith(authentication).
			Handle(promhttp.Handler()),

		route.Post("/accounts/import").
			SecuredWith(authentication).
			Handle(handlers.PostAccountsImport(app)),

		route.Get("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(handlers.GetAccount(app)),

		route.Patch("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(handlers.PatchAccount(app)),

		route.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(handlers.PatchAccountLock(app)),

		route.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(handlers.PatchAccountUnlock(app)),

		route.Patch("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(handlers.PatchAccountExpirePassword(app)),

		route.Put("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(handlers.PatchAccount(app)),

		route.Put("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(handlers.PatchAccountLock(app)),

		route.Put("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(handlers.PatchAccountUnlock(app)),

		route.Put("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(handlers.PatchAccountExpirePassword(app)),

		route.Delete("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(handlers.DeleteAccount(app)),

		route.Delete("/accounts/{id:[0-9]+}/oauth/{name}").
			SecuredWith(authentication).
			Handle(handlers.DeleteAccountOauth(app)),
	)

	if app.Actives != nil {
		routes = append(routes,
			route.Get("/stats").
				SecuredWith(authentication).
				Handle(handlers.GetStats(app)),
		)
	}

	return routes
}
