package main

import (
	"log"
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn/handlers"
)

func main() {
	// set up connections and configuration
	app, err := handlers.NewApp()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", app.Stub).Methods("GET")

	attach(r,
		post("/accounts").
			securedWith(handlers.RefererSecurity(app.Config.ApplicationDomains)).
			handle(app.PostAccount),
	)
	r.HandleFunc("/accounts/import", app.Stub).Methods("POST")
	r.HandleFunc("/accounts/available", app.Stub).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", app.Stub).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", app.Stub).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", app.Stub).Methods("PUT", "PATCH")

	r.HandleFunc("/session", app.Stub).Methods("POST")
	r.HandleFunc("/session", app.Stub).Methods("DELETE")
	r.HandleFunc("/session/refresh", app.Stub).Methods("GET")

	r.HandleFunc("/password", app.Stub).Methods("POST")
	r.HandleFunc("/password/reset", app.Stub).Methods("GET")

	r.HandleFunc("/configuration", app.Stub).Methods("GET")
	r.HandleFunc("/jwks", app.Stub).Methods("GET")

	r.HandleFunc("/stats", app.Stub).Methods("GET")
	r.HandleFunc("/health", app.Health).Methods("GET")

	corsAdapter := gorilla.CORS(
		gorilla.AllowedOrigins(app.Config.ApplicationOrigins),
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
	)
	recoveryAdapter := gorilla.RecoveryHandler()

	stack := recoveryAdapter(
		corsAdapter(
			gorilla.CombinedLoggingHandler(os.Stdout, r),
		),
	)

	log.Fatal(http.ListenAndServe(":8000", stack))
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
