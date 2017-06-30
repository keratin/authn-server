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

	// TODO: MountedPath
	attach(r,
		post("/accounts").
			securedWith(refererSecurity).
			handle(accounts.PostAccount(app)),
	)
	r.HandleFunc("/accounts/import", api.Stub(app)).Methods("POST")
	r.HandleFunc("/accounts/available", api.Stub(app)).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", api.Stub(app)).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", api.Stub(app)).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", api.Stub(app)).Methods("PUT", "PATCH")

	attach(r,
		post("/session").
			securedWith(refererSecurity).
			handle(sessions.PostSession(app)),
	)
	attach(r,
		delete("/session").
			securedWith(refererSecurity).
			handle(sessions.DeleteSession(app)),
	)
	attach(r,
		get("/session/refresh").
			securedWith(refererSecurity).
			handle(sessions.GetSessionRefresh(app)),
	)

	r.HandleFunc("/password", api.Stub(app)).Methods("POST")
	r.HandleFunc("/password/reset", api.Stub(app)).Methods("GET")

	r.HandleFunc("/configuration", api.Stub(app)).Methods("GET")
	r.HandleFunc("/jwks", api.Stub(app)).Methods("GET")

	r.HandleFunc("/stats", api.Stub(app)).Methods("GET")
	r.HandleFunc("/health", health.Health(app)).Methods("GET")

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

type route struct {
	verb string
	tpl  string
}

type securedRoute struct {
	route    *route
	security api.SecurityHandler
}

type handledRoute struct {
	route   *securedRoute
	handler http.Handler
}

func post(tpl string) *route {
	return &route{verb: "POST", tpl: tpl}
}

func get(tpl string) *route {
	return &route{verb: "GET", tpl: tpl}
}

func delete(tpl string) *route {
	return &route{verb: "DELETE", tpl: tpl}
}

func (r *route) securedWith(fn api.SecurityHandler) *securedRoute {
	return &securedRoute{route: r, security: fn}
}

func (r *securedRoute) handle(fn func(w http.ResponseWriter, r *http.Request)) *handledRoute {
	return &handledRoute{route: r, handler: http.HandlerFunc(fn)}
}

func attach(router *mux.Router, r *handledRoute) {
	router.
		Methods(r.route.route.verb).
		Path(r.route.route.tpl).
		Handler(r.route.security(r.handler))
}
