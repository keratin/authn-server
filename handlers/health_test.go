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

	handler := http.HandlerFunc(handlers.Health)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusOK)
	AssertBody(t, res, `{"http":true}`)
}
