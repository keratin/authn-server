package public

import (
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/oauth"
	"github.com/keratin/authn-server/lib/route"
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
	route.Attach(router, app.Config.MountedPath, accountRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, metaRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, sessionsRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, passwordsRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, oauth.PublicRoutes(app)...)
}

func metaRoutes(app *api.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	return []*route.HandledRoute{
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(gmux),
	}
}

func accountRoutes(app *api.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{}

	if app.Config.EnableSignup {
		routes = append(routes,
			route.Post("/accounts").
				SecuredWith(originSecurity).
				Handle(gmux),
			route.Get("/accounts/available").
				SecuredWith(originSecurity).
				Handle(gmux),
		)
	}

	return routes
}

func sessionsRoutes(app *api.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/session").
			SecuredWith(originSecurity).
			Handle(gmux),

		route.Delete("/session").
			SecuredWith(originSecurity).
			Handle(gmux),

		route.Get("/session/refresh").
			SecuredWith(originSecurity).
			Handle(gmux),
	}

	if app.Config.AppPasswordlessTokenURL != nil {
		routes = append(routes,
			route.Get("/session/token").
				SecuredWith(originSecurity).
				Handle(gmux),

			route.Post("/session/token").
				SecuredWith(originSecurity).
				Handle(gmux),
		)
	}

	return routes
}

func passwordsRoutes(app *api.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	routes := []*route.HandledRoute{
		route.Post("/password").
			SecuredWith(originSecurity).
			Handle(gmux),
	}

	if app.Config.AppPasswordResetURL != nil {
		routes = append(routes,
			route.Get("/password/reset").
				SecuredWith(originSecurity).
				Handle(gmux),
		)
	}

	return routes
}
