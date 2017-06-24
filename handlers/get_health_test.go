package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn-server/handlers"
)

func TestHealth(t *testing.T) {
	app := handlers.App{
		DbCheck:    func() bool { return true },
		RedisCheck: func() bool { return true },
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	assertCode(t, res, http.StatusOK)
	assertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
