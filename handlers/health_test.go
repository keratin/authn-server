package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn/handlers"
)

const HTTP_STATUS_FAILURE = "HTTP status:\n  expected: %v\n  actual:   %v"
const HTTP_BODY_FAILURE = "HTTP body:\n  expected: %v\n  actual:   %v"

func TestHealth(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(handlers.Health)
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf(HTTP_STATUS_FAILURE, http.StatusOK, status)
	}

	expected := "up"
	if res.Body.String() != expected {
		t.Errorf(HTTP_BODY_FAILURE, expected, res.Body.String())
	}
}
