package passwords

import (
	"github.com/keratin/authn-server/api/handlers"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/password").
			SecuredWith(originSecurity).
			Handle(handlers.PostPassword(app)),
	}

	if app.Config.AppPasswordResetURL != nil {
		routes = append(routes,
			route.Get("/password/reset").
				SecuredWith(originSecurity).
				Handle(handlers.GetPasswordReset(app)),
		)
	}

	return routes
}

func Routes(app *app.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
