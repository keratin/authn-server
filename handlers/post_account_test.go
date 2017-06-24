package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keratin/authn-server/services"
)

func TestPostAccountSuccess(t *testing.T) {
	app := testApp()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("username=foo&password=bar"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	assertCode(t, res, http.StatusCreated)
	assertSession(t, res)
	assertIdTokenResponse(t, res, app.Config)
}

func TestPostAccountSuccessWithSession(t *testing.T) {
	app := testApp()

	account_id := 8642

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("username=foo&password=bar"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(createSession(t, app.RefreshTokenStore, app.Config, account_id))

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(account_id)
	if err != nil {
		t.Fatal(err)
	}
	refreshToken := refreshTokens[0]

	handler := http.HandlerFunc(app.PostAccount)
	handler.ServeHTTP(res, req)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if id != 0 {
		t.Errorf("Expected token to be revoked: %v", refreshToken)
	}
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
