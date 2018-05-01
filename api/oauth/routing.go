package oauth

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *api.App) []*route.HandledRoute {
	// TODO: originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	var routes []*route.HandledRoute

	for providerName := range app.OauthProviders {
		routes = append(routes,
			route.Get("/oauth/"+providerName).
				SecuredWith(route.Unsecured()).
				Handle(getOauth(app, providerName)),
			route.Get("/oauth/"+providerName+"/return").
				SecuredWith(route.Unsecured()).
				Handle(getOauthReturn(app, providerName)),
		)
	}

	return routes
}

func Routes(app *api.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
