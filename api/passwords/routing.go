package passwords

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func Routes(app *api.App) []*route.HandledRoute {
	refererSecurity := route.RefererSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/password").
			SecuredWith(refererSecurity).
			Handle(postPassword(app)),
	}

	if app.Config.AppPasswordResetURL != nil {
		routes = append(routes,
			route.Get("/password/reset").
				SecuredWith(refererSecurity).
				Handle(getPasswordReset(app)),
		)
	}

	return routes
}
