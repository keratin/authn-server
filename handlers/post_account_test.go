package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type dict map[string]string

func TestPostAccountSuccess(t *testing.T) {
	app := App()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("username=foo&password=bar"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusCreated)
	AssertResult(t, res, dict{"id_token": "j.w.t"})
}

func TestPostAccountFailure(t *testing.T) {
	app := App()
	handler := http.HandlerFunc(app.PostAccount)

	var tests = []struct {
		body   string
		errors dict
	}{
		{"", dict{"foo": "bar"}},
	}

	for _, tt := range tests {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/accounts", strings.NewReader(tt.body))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		handler.ServeHTTP(res, req)

		AssertCode(t, res, http.StatusUnprocessableEntity)
		AssertErrors(t, res, tt.errors)
	}
}
