package passwords

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *api.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/password").
			SecuredWith(originSecurity).
			Handle(postPassword(app)),
	}

	if app.Config.AppPasswordResetURL != nil {
		routes = append(routes,
			route.Get("/password/reset").
				SecuredWith(originSecurity).
				Handle(getPasswordReset(app)),
		)
	}

	return routes
}

func Routes(app *api.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
