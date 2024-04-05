package server

import (
	"net/http"

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

		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetJWKs(app)),

		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetConfiguration(app)),

		route.Post("/password").
			SecuredWith(originSecurity).
			Handle(handlers.PostPassword(app)),

		route.Post("/password/score").
			SecuredWith(originSecurity).
			Handle(handlers.PostPasswordScore(app)),

		route.Post("/session").
			SecuredWith(originSecurity).
			Handle(handlers.PostSession(app)),

		route.Delete("/session").
			SecuredWith(originSecurity).
			Handle(handlers.DeleteSession(app)),

		route.Get("/session/refresh").
			SecuredWith(originSecurity).
			Handle(handlers.GetSessionRefresh(app)),

		route.Post("/totp/new").
			SecuredWith(originSecurity).
			Handle(handlers.CreateTOTP(app)),

		route.Post("/totp/confirm").
			SecuredWith(originSecurity).
			Handle(handlers.ConfirmTOTP(app)),

		route.Delete("/totp").
			SecuredWith(originSecurity).
			Handle(handlers.DeleteTOTP(app)),

		route.Get("/oauth/accounts").
			SecuredWith(originSecurity).
			Handle(handlers.GetOauthAccounts(app)),
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

	for providerName, provider := range app.OauthProviders {
		var returnRoute *route.Route
		if provider.ReturnMethod() == http.MethodPost {
			returnRoute = route.Post("/oauth/" + providerName + "/return")
		} else {
			returnRoute = route.Get("/oauth/" + providerName + "/return")
		}
		routes = append(routes,
			route.Get("/oauth/"+providerName).
				SecuredWith(route.Unsecured()).
				Handle(handlers.GetOauth(app, providerName)),
			returnRoute.
				SecuredWith(route.Unsecured()).
				Handle(handlers.GetOauthReturn(app, providerName)),
			route.Delete("/oauth/"+providerName).
				SecuredWith(originSecurity).
				Handle(handlers.DeleteOauth(app, providerName)),
		)
	}

	return routes
}
