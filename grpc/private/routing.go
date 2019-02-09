package private

import (
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/grpc/public"
	"github.com/keratin/authn-server/lib/route"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
func router(app *api.App) http.Handler {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, meta.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, oauth.Routes(app)...)

	return wrapRouter(r, app)
}

func publicRouter(app *api.App) http.Handler {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, meta.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, oauth.PublicRoutes(app)...)

	return wrapRouter(r, app)
}
*/

func RegisterRoutes(router *mux.Router, app *api.App, gmux *runtime.ServeMux) {
	public.RegisterRoutes(router, app, gmux)

	route.Attach(router, app.Config.MountedPath, metaRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, accountRoutes(app, gmux)...)

}

func metaRoutes(app *api.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := []*route.HandledRoute{}

	if app.Actives != nil {
		routes = append(routes,
			route.Get("/stats").
				SecuredWith(authentication).
				Handle(gmux),
		)
	}

	routes = append(routes,
		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(gmux),
		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(gmux),
		route.Get("/metrics").
			SecuredWith(authentication).
			Handle(promhttp.Handler()),
	)

	return routes
}

func accountRoutes(app *api.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	routes := []*route.HandledRoute{}

	routes = append(routes,
		route.Post("/accounts/import").
			SecuredWith(authentication).
			Handle(gmux),

		route.Get("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gmux),

		route.Patch("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gmux),

		route.Patch("/accounts/{id:[0-9]+}/lock").
			SecuredWith(authentication).
			Handle(gmux),

		route.Patch("/accounts/{id:[0-9]+}/unlock").
			SecuredWith(authentication).
			Handle(gmux),

		route.Patch("/accounts/{id:[0-9]+}/expire_password").
			SecuredWith(authentication).
			Handle(gmux),

		route.Delete("/accounts/{id:[0-9]+}").
			SecuredWith(authentication).
			Handle(gmux),
	)

	return routes
}
