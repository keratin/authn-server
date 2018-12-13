package meta

import (
	"github.com/keratin/authn-server/api/handlers"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {
	return []*route.HandledRoute{
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetHealth(app)),
	}
}

func Routes(app *app.App) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := PublicRoutes(app)

	if app.Actives != nil {
		routes = append(routes,
			route.Get("/stats").
				SecuredWith(authentication).
				Handle(handlers.GetStats(app)),
		)
	}

	routes = append(routes,
		route.Get("/").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetRoot(app)),
		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetJWKs(app)),
		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetConfiguration(app)),
		route.Get("/metrics").
			SecuredWith(authentication).
			Handle(promhttp.Handler()),
	)

	return routes
}
