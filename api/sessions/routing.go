package sessions

import (
	"github.com/keratin/authn-server/api/handlers"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/session").
			SecuredWith(originSecurity).
			Handle(handlers.PostSession(app)),

		route.Delete("/session").
			SecuredWith(originSecurity).
			Handle(handlers.DeleteSession(app)),

		route.Get("/session/refresh").
			SecuredWith(originSecurity).
			Handle(handlers.GetSessionRefresh(app)),
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

	return routes
}

func Routes(app *app.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
