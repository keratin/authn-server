package server

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/handlers"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {
	var routes []*route.HandledRoute
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains, app.Logger)

	routes = append(routes,
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetHealth(app)),

		route.Post("/password").
			SecuredWith(originSecurity).
			Handle(handlers.PostPassword(app)),

		route.Post("/session").
			SecuredWith(originSecurity).
			Handle(handlers.PostSession(app)),

		route.Delete("/session").
			SecuredWith(originSecurity).
			Handle(handlers.DeleteSession(app)),

		route.Get("/session/refresh").
			SecuredWith(originSecurity).
			Handle(handlers.GetSessionRefresh(app)),
	)

	if app.Config.EnableSignup {
		routes = append(routes,
			route.Post("/accounts").
				SecuredWith(originSecurity).
				Handle(handlers.PostAccount(app)),
			route.Get("/accounts/available").
				SecuredWith(originSecurity).
				Handle(handlers.GetAccountsAvailable(app)),
		)
	}

	if app.Config.AppPasswordResetURL != nil {
		routes = append(routes,
			route.Get("/password/reset").
				SecuredWith(originSecurity).
				Handle(handlers.GetPasswordReset(app)),
		)
	}

	if app.Config.AppPasswordlessTokenURL != nil {
		routes = append(routes,
			route.Get("/session/token").
				SecuredWith(originSecurity).
				Handle(handlers.GetSessionToken(app)),

			route.Post("/session/token").
				SecuredWith(originSecurity).
				Handle(handlers.PostSessionToken(app)),
		)
	}

	for providerName := range app.OauthProviders {
		routes = append(routes,
			route.Get("/oauth/"+providerName).
				SecuredWith(route.Unsecured()).
				Handle(handlers.GetOauth(app, providerName)),
			route.Get("/oauth/"+providerName+"/return").
				SecuredWith(route.Unsecured()).
				Handle(handlers.GetOauthReturn(app, providerName)),
		)
	}

	return routes
}
