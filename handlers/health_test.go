package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn/handlers"
)

func TestHealth(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	app := handlers.App{"test"}

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusOK)
	AssertBody(t, res, `{"http":true}`)
}
