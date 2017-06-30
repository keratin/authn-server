package main

import (
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/handlers"
)

func routing(app *handlers.App) http.Handler {
	r := mux.NewRouter()

	refererSecurity := handlers.RefererSecurity(app.Config.ApplicationDomains)

	r.HandleFunc("/", handlers.Stub(app)).Methods("GET")

	// TODO: MountedPath
	attach(r,
		post("/accounts").
			securedWith(refererSecurity).
			handle(handlers.PostAccount(app)),
	)
	r.HandleFunc("/accounts/import", handlers.Stub(app)).Methods("POST")
	r.HandleFunc("/accounts/available", handlers.Stub(app)).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", handlers.Stub(app)).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", handlers.Stub(app)).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", handlers.Stub(app)).Methods("PUT", "PATCH")

	attach(r,
		post("/session").
			securedWith(refererSecurity).
			handle(handlers.PostSession(app)),
	)
	attach(r,
		delete("/session").
			securedWith(refererSecurity).
			handle(handlers.Stub(app)),
	)
	attach(r,
		get("/session/refresh").
			securedWith(refererSecurity).
			handle(handlers.GetSessionRefresh(app)),
	)

	r.HandleFunc("/password", handlers.Stub(app)).Methods("POST")
	r.HandleFunc("/password/reset", handlers.Stub(app)).Methods("GET")

	r.HandleFunc("/configuration", handlers.Stub(app)).Methods("GET")
	r.HandleFunc("/jwks", handlers.Stub(app)).Methods("GET")

	r.HandleFunc("/stats", handlers.Stub(app)).Methods("GET")
	r.HandleFunc("/health", handlers.Health(app)).Methods("GET")

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
	security handlers.SecurityHandler
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

func (r *route) securedWith(fn handlers.SecurityHandler) *securedRoute {
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
