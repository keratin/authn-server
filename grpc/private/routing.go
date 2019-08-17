package private

import (
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/grpc/internal/gateway"
	"github.com/keratin/authn-server/grpc/public"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
	Reference: github.com/keratin/authn-server/server/private_routes.go
*/

// RegisterRoutes registers gmux as the handler for the private routes on router
func registerRoutes(router *mux.Router, app *app.App, gmux *runtime.ServeMux) {
	public.RegisterRoutes(router, app, gmux)

	route.Attach(router, app.Config.MountedPath, metaRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, accountRoutes(app, gmux)...)

}

func metaRoutes(app *app.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := []*route.HandledRoute{}

	if app.Actives != nil {
		routes = append(routes,
			route.Get("/stats").
				SecuredWith(authentication).
				Handle(gateway.TrimSubpath(app, gmux)),
		)
	}

	routes = append(routes,
		route.Get("/").
			SecuredWith(route.Unsecured()).
			Handle(handlers.GetRoot(app)),
		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(gateway.TrimSubpath(app, gmux)),
		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(gateway.TrimSubpath(app, gmux)),
		route.Get("/metrics").
			SecuredWith(authentication).
			Handle(promhttp.Handler()),
	)

	return routes
}

func accountRoutes(app *app.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := []*route.HandledRoute{}

	routes = append(routes,
		route.Post("/accounts/import").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Get("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Patch("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Patch("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Put("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Put("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Put("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Put("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),

		route.Delete("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gateway.TrimSubpath(app, gmux)),
	)

	return routes
}
