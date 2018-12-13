package test

import (
	"net/http/httptest"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/app"
)

func Server(app *app.App) *httptest.Server {
	return httptest.NewServer(api.Router(app))
}
