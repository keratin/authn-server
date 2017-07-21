package test

import (
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
)

func Server(app *api.App, routes []*api.HandledRoute) *httptest.Server {
	r := mux.NewRouter()
	api.Attach(r, app.Config.MountedPath, routes...)
	return httptest.NewServer(api.Session(app)(r))
}
