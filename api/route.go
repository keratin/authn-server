package api

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

func Post(tpl string) *route {
	return &route{verb: "POST", tpl: tpl}
}

func Get(tpl string) *route {
	return &route{verb: "GET", tpl: tpl}
}

func Delete(tpl string) *route {
	return &route{verb: "DELETE", tpl: tpl}
}

type route struct {
	verb string
	tpl  string
}

func (r *route) SecuredWith(fn SecurityHandler) *securedRoute {
	return &securedRoute{route: r, security: fn}
}

type securedRoute struct {
	route    *route
	security SecurityHandler
}

func (r *securedRoute) Handle(fn func(w http.ResponseWriter, r *http.Request)) *handledRoute {
	return &handledRoute{route: r, handler: http.HandlerFunc(fn)}
}

type handledRoute struct {
	route   *securedRoute
	handler http.Handler
}

func Attach(router *mux.Router, pathPrefix string, routes ...*handledRoute) {
	re := regexp.MustCompile("/{2,}")

	for _, r := range routes {
		path := re.ReplaceAllLiteralString(
			fmt.Sprintf("/%s/%s", pathPrefix, r.route.route.tpl),
			"/",
		)

		router.
			Methods(r.route.route.verb).
			Path(path).
			Handler(r.route.security(r.handler))
	}
}
