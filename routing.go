package main

import (
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/health"
	"github.com/keratin/authn-server/api/sessions"
)

func router(app *api.App) http.Handler {
	r := mux.NewRouter()

	// GET  /
	// POST /password
	// GET  /password/reset
	// GET  /configuration
	// GET  /jwks
	// GET  /stats

	api.Attach(r, app.Config.MountedPath, health.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)

	corsAdapter := gorilla.CORS(
		gorilla.AllowedOrigins(app.Config.ApplicationOrigins),
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
	)

	return gorilla.RecoveryHandler()(
		corsAdapter(
			gorilla.CombinedLoggingHandler(os.Stdout, r),
		),
	)
}
