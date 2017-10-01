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
)

func router(app *api.App) http.Handler {
	r := mux.NewRouter()

	route.Attach(r, app.Config.MountedPath, meta.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.Routes(app)...)

	corsAdapter := gorilla.CORS(
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		gorilla.AllowCredentials(),
		gorilla.AllowedOriginValidator(api.OriginValidator(app.Config.ApplicationDomains)),
	)

	session := api.Session(app)

	return app.Reporter.PanicHandler(
		corsAdapter(
			session(
				gorilla.CombinedLoggingHandler(os.Stdout, r),
			),
		),
	)
}
