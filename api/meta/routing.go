package meta

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Routes(app *api.App) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := []*route.HandledRoute{}

	if app.Actives != nil {
		routes = append(routes,
			route.Get("/stats").
				SecuredWith(authentication).
				Handle(getStats(app)),
		)
	}

	routes = append(routes,
		route.Get("/").
			SecuredWith(route.Unsecured()).
			Handle(getRoot(app)),
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(getHealth(app)),
		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(getJWKs(app)),
		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(getConfiguration(app)),
		route.Get("/metrics").
			SecuredWith(authentication).
			Handle(promhttp.Handler()),
	)

	return routes
}
