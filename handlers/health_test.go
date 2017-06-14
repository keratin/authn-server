package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	app := testApp()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	assertCode(t, res, http.StatusOK)
	assertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
