package public

import (
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/handlers"
)

/*
	Reference: github.com/keratin/authn-server/server/public_routes.go
*/

// RegisterRoutes registers gmux as the handler for the public routes on router
func RegisterRoutes(router *mux.Router, app *app.App, gmux *runtime.ServeMux) {
	route.Attach(router, app.Config.MountedPath, accountRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, metaRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, sessionsRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, passwordsRoutes(app, gmux)...)
	route.Attach(router, app.Config.MountedPath, oauthRoutes(app)...)
}

func metaRoutes(app *app.App, gmux *runtime.ServeMux) []*route.HandledRoute {
	return []*route.HandledRoute{
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(gmux),
	}
}

func accountRoutes(app *app.App, gmux *runtime.ServeMux) []*route.HandledRoute {
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

func sessionsRoutes(app *app.App, gmux *runtime.ServeMux) []*route.HandledRoute {
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

func passwordsRoutes(app *app.App, gmux *runtime.ServeMux) []*route.HandledRoute {
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

func oauthRoutes(app *app.App) []*route.HandledRoute {
	routes := []*route.HandledRoute{}
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
