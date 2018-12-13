package oauth

import (
	"github.com/keratin/authn-server/api/handlers"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {

	var routes []*route.HandledRoute

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

func Routes(app *app.App) []*route.HandledRoute {
	return PublicRoutes(app)
}
