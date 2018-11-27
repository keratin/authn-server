package sessions

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *api.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/session").
			SecuredWith(originSecurity).
			Handle(postSession(app)),

		route.Delete("/session").
			SecuredWith(originSecurity).
			Handle(deleteSession(app)),

		route.Get("/session/refresh").
			SecuredWith(originSecurity).
			Handle(getSessionRefresh(app)),
	}

	if app.Config.AppPasswordlessTokenURL != nil {
		routes = append(routes,
			route.Get("/session/token").
				SecuredWith(originSecurity).
				Handle(getSessionToken(app)),

			route.Post("/session/token").
				SecuredWith(originSecurity).
				Handle(postSessionToken(app)),
		)
	}

	return routes
}

func Routes(app *api.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
