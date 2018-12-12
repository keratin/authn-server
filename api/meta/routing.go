package meta

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PublicRoutes(app *app.App) []*route.HandledRoute {
	return []*route.HandledRoute{
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(GetHealth(app)),
	}
}

func Routes(app *app.App) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := PublicRoutes(app)

	if app.Actives != nil {
		routes = append(routes,
			route.Get("/stats").
				SecuredWith(authentication).
				Handle(GetStats(app)),
		)
	}

	routes = append(routes,
		route.Get("/").
			SecuredWith(route.Unsecured()).
			Handle(GetRoot(app)),
		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(GetJWKs(app)),
		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(GetConfiguration(app)),
		route.Get("/metrics").
			SecuredWith(authentication).
			Handle(promhttp.Handler()),
	)

	return routes
}
