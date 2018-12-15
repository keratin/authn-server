package test

import (
	"net/http/httptest"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/server"
)

func Server(app *app.App) *httptest.Server {
	return httptest.NewServer(server.Router(app))
}
