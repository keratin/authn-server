package passwordless

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *api.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

  routes := []*route.HandledRoute{}

	if app.Config.AppPasswordlessTokenURL != nil {
		routes = append(routes,
			route.Get("/passwordless/token").
				SecuredWith(originSecurity).
				Handle(getPasswordlessToken(app)),
		)
	}

	return routes
}

func Routes(app *api.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
