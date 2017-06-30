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

	res := get("/health", handlers.Health(&app))

	assertCode(t, res, http.StatusOK)
	assertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
