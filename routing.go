package main

import (
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/meta"
	"github.com/keratin/authn-server/api/passwords"
	"github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
)

func router(app *api.App) http.Handler {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, meta.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.Routes(app)...)

	return wrapRouter(r, app)
}

func publicRouter(app *api.App) http.Handler {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, meta.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.PublicRoutes(app)...)

	return wrapRouter(r, app)
}

func wrapRouter(r *mux.Router, app *api.App) http.Handler {
	corsAdapter := gorilla.CORS(
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		gorilla.AllowCredentials(),
		gorilla.AllowedOrigins([]string{}), // see: https://github.com/gorilla/handlers/issues/117
		gorilla.AllowedOriginValidator(api.OriginValidator(app.Config.ApplicationDomains)),
	)

	session := api.Session(app)

	return ops.PanicHandler(app.Reporter,
		corsAdapter(
			session(
				gorilla.CombinedLoggingHandler(os.Stdout, r),
			),
		),
	)
}
