package test

import (
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func Server(app *api.App, routes []*route.HandledRoute) *httptest.Server {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, routes...)
	return httptest.NewServer(api.Session(app)(r))
}
