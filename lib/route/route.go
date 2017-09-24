// Package route is a fluent API for building secured gorilla/mux routes. It requires every route to
// have a defined SecurityHandler, which is a middleware that encodes some authorization strategy.
// Included SecurityHandlers can authorize a request by HTTP Referer (for CSRF security) or HTTP
// Basic Auth. There is also an Unsecured handler for explicitly acknowledging that an endpoint is
// wide open.
package route

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SecurityHandler is a kind of middleware that satisfies security requirements
type SecurityHandler func(http.Handler) http.Handler

// Post creates a new POST route. A security handler must be registered next.
func Post(tpl string) *route {
	return &route{verb: "POST", tpl: tpl}
}

// Get creates a new GET route. A security handler must be registered next.
func Get(tpl string) *route {
	return &route{verb: "GET", tpl: tpl}
}

// Delete creates a new DELETE route. A security handler must be registered next.
func Delete(tpl string) *route {
	return &route{verb: "DELETE", tpl: tpl}
}

// Patch creates a new PATCH route. A security handler must be registered next.
func Patch(tpl string) *route {
	return &route{verb: "PATCH", tpl: tpl}
}

// route is private because it is incomplete.
type route struct {
	verb string
	tpl  string
}

// SecuredWith registers a security handler for a route. A handler must be registered next.
func (r *route) SecuredWith(fn SecurityHandler) *securedRoute {
	return &securedRoute{route: r, security: fn}
}

// securedRoute is private because it is incomplete.
type securedRoute struct {
	route    *route
	security SecurityHandler
}

// Handle registers a function as a HandlerFunc. The route may now be attached.
func (r *securedRoute) Handle(fn func(w http.ResponseWriter, r *http.Request)) *HandledRoute {
	return &HandledRoute{route: r, handler: http.HandlerFunc(fn)}
}

// HandledRoute is a fully defined, attachable route.
type HandledRoute struct {
	route   *securedRoute
	handler http.Handler
}

// Attach is the adapter for adding HandledRoutes to a gorilla/mux Router.
func Attach(router *mux.Router, pathPrefix string, routes ...*HandledRoute) {
	for _, r := range routes {
		router.
			PathPrefix(pathPrefix).
			Methods(r.route.route.verb).
			Path(r.route.route.tpl).
			Handler(r.route.security(r.handler))
	}
}
