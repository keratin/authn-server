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
	return &Route{Verb: "POST", Tpl: tpl}
}

// Get creates a new GET route. A security handler must be registered next.
func Get(tpl string) *Route {
	return &Route{Verb: "GET", Tpl: tpl}
}

// Delete creates a new DELETE route. A security handler must be registered next.
func Delete(tpl string) *Route {
	return &Route{Verb: "DELETE", Tpl: tpl}
}

// Patch creates a new PATCH route. A security handler must be registered next.
func Patch(tpl string) *Route {
	return &Route{Verb: "PATCH", Tpl: tpl}
}

// Put creates a new PUT route. A security handler must be registered next.
func Put(tpl string) *Route {
	return &Route{Verb: "PUT", Tpl: tpl}
}

// Route is an incomplete Route comprising only verb and path (as a gorilla/mux template). It must
// next be `SecuredWith`.
type Route struct {
	Verb string
	Tpl  string
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

func (hr *HandledRoute) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	hr.security(hr.handler).ServeHTTP(response, request)
}

// Attach is the adapter for adding HandledRoutes to a gorilla/mux Router.
func Attach(router *mux.Router, pathPrefix string, routes ...*HandledRoute) {
	for _, r := range routes {
		router.
			PathPrefix(pathPrefix).
			Methods(r.Verb).
			Path(r.Tpl).
			Handler(InstrumentRoute(r.Verb+" "+r.Tpl, r))
	}
}
