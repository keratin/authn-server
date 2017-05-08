package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostAccountSuccess(t *testing.T) {
	app := App()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("body"))

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusCreated)
	AssertBody(t, res, `{"result":{"id_token":"j.w.t"}}`)
}

func TestPostAccountFailure(t *testing.T) {
	app := App()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", nil)

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusUnprocessableEntity)
	AssertBody(t, res, `{"errors":[{"field":"foo","message":"bar"}]}`)
}
