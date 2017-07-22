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
)

func router(app *api.App) http.Handler {
	r := mux.NewRouter()

	// GET  /
	// GET  /configuration
	// GET  /jwks
	// GET  /stats

	api.Attach(r, app.Config.MountedPath, meta.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, passwords.Routes(app)...)

	corsAdapter := gorilla.CORS(
		gorilla.AllowedOrigins(app.Config.ApplicationOrigins),
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
	)

	session := api.Session(app)

	return gorilla.RecoveryHandler()(
		corsAdapter(
			session(
				gorilla.CombinedLoggingHandler(os.Stdout, r),
			),
		),
	)
}
