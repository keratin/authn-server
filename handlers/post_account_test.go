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

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusUnprocessableEntity)
	AssertErrors(t, res, dict{"foo": "bar"})
}
