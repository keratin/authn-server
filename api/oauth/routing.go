package oauth

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *api.App) []*route.HandledRoute {
	// originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	return []*route.HandledRoute{
		route.Get("/oauth/{provider}").
			SecuredWith(route.Unsecured()).
			Handle(startOauth(app)),
		route.Get("/oauth/{provider}/return").
			SecuredWith(route.Unsecured()).
			Handle(completeOauth(app)),
	}
}

func Routes(app *api.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
