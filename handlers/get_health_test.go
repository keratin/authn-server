package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/handlers"
)

func TestHealth(t *testing.T) {
	app := handlers.App{
		DbCheck:    func() bool { return true },
		RedisCheck: func() bool { return true },
	}

	res := get("/health", app.Health)

	assertCode(t, res, http.StatusOK)
	assertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
