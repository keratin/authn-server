package route_test

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/lib/route"
)

func ExampleRoute() {
	r := mux.NewRouter()
	basicAuth := route.BasicAuthSecurity("username", "password", "Realm Name")

	privateHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	route.Attach(r, "/",
		route.Get("/private").
			SecuredWith(basicAuth).
			Handle(privateHandler),

		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(healthHandler),
	)
}
