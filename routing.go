package main

import (
	"net/http"
	"net/url"
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

	api.Attach(r, app.Config.MountedPath, meta.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)
	api.Attach(r, app.Config.MountedPath, passwords.Routes(app)...)

	corsAdapter := gorilla.CORS(
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		gorilla.AllowCredentials(),
		gorilla.AllowedOriginValidator(func(origin string) bool {
			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}

			for _, appDomain := range app.Config.ApplicationDomains {
				if appDomain.Matches(originURL) {
					return true
				}
			}
			return false
		}),
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
