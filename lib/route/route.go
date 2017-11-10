// Package route is a fluent API for building secured gorilla/mux routes. It requires every route to
// have a defined SecurityHandler, which is a middleware that encodes some authorization strategy.
// Included SecurityHandlers can authorize a request by HTTP Origin (for CSRF security) or HTTP
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
func Post(tpl string) *Route {
	return &Route{verb: "POST", tpl: tpl}
}

// Get creates a new GET route. A security handler must be registered next.
func Get(tpl string) *Route {
	return &Route{verb: "GET", tpl: tpl}
}

// Delete creates a new DELETE route. A security handler must be registered next.
func Delete(tpl string) *Route {
	return &Route{verb: "DELETE", tpl: tpl}
}

// Patch creates a new PATCH route. A security handler must be registered next.
func Patch(tpl string) *Route {
	return &Route{verb: "PATCH", tpl: tpl}
}

// Route is an incomplete Route comprising only verb and path (as a gorilla/mux template). It must
// next be `SecuredWith`.
type Route struct {
	verb string
	tpl  string
}

// SecuredWith registers a security handler for a route. A handler must be registered next.
func (r Route) SecuredWith(fn SecurityHandler) *SecuredRoute {
	return &SecuredRoute{r, fn}
}

// SecuredRoute is an incomplete Route with a defined SecurityHandler. It is ready for a
// http.Handler.
type SecuredRoute struct {
	Route
	security SecurityHandler
}

// Handle registers a HandlerFunc. The route may now be `Attach`d.
func (r *SecuredRoute) Handle(h http.Handler) *HandledRoute {
	return &HandledRoute{r, h}
}

// HandledRoute is a fully defined route. It is ready to be `Attach`d.
type HandledRoute struct {
	*SecuredRoute
	handler http.Handler
}

// Attach is the adapter for adding HandledRoutes to a gorilla/mux Router.
func Attach(router *mux.Router, pathPrefix string, routes ...*HandledRoute) {
	for _, r := range routes {
		router.
			PathPrefix(pathPrefix).
			Methods(r.verb).
			Path(r.tpl).
			Handler(r.security(r.handler))
	}
}
