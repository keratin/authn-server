package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keratin/authn/services"
)

func TestPostAccountSuccess(t *testing.T) {
	app := testApp()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("username=foo&password=bar"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	assertCode(t, res, http.StatusCreated)
	assertResult(t, res, map[string]string{"id_token": "j.w.t"})
}

func TestPostAccountFailure(t *testing.T) {
	app := testApp()
	handler := http.HandlerFunc(app.PostAccount)

	var tests = []struct {
		body   string
		errors []services.Error
	}{
		{"", []services.Error{{"username", "MISSING"}, {"password", "MISSING"}}},
	}

	for _, tt := range tests {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/accounts", strings.NewReader(tt.body))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		handler.ServeHTTP(res, req)

		assertCode(t, res, http.StatusUnprocessableEntity)
		assertErrors(t, res, tt.errors)
	}
}
