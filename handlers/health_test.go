package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	app := App()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusOK)
	AssertBody(t, res, `{"http":true,"db":true}`)
}
