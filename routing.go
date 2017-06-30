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

func routing(app *api.App) http.Handler {
	r := mux.NewRouter()

	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)

	r.HandleFunc("/", api.Stub(app)).Methods("GET")

	api.Attach(r, app.Config.MountedPath,
		api.Post("/accounts").
			SecuredWith(refererSecurity).
			Handle(accounts.PostAccount(app)),
	)
	r.HandleFunc("/accounts/import", api.Stub(app)).Methods("POST")
	r.HandleFunc("/accounts/available", api.Stub(app)).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", api.Stub(app)).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", api.Stub(app)).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", api.Stub(app)).Methods("PUT", "PATCH")

	api.Attach(r, app.Config.MountedPath,
		api.Post("/session").
			SecuredWith(refererSecurity).
			Handle(sessions.PostSession(app)),
		api.Delete("/session").
			SecuredWith(refererSecurity).
			Handle(sessions.DeleteSession(app)),
		api.Get("/session/refresh").
			SecuredWith(refererSecurity).
			Handle(sessions.GetSessionRefresh(app)),
	)

	r.HandleFunc("/password", api.Stub(app)).Methods("POST")
	r.HandleFunc("/password/reset", api.Stub(app)).Methods("GET")

	r.HandleFunc("/configuration", api.Stub(app)).Methods("GET")
	r.HandleFunc("/jwks", api.Stub(app)).Methods("GET")

	r.HandleFunc("/stats", api.Stub(app)).Methods("GET")
	api.Attach(r, app.Config.MountedPath,
		api.Get("/health").
			SecuredWith(api.Unsecured()).
			Handle(health.GetHealth(app)),
	)

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
