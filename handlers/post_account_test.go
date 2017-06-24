package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/services"
)

func TestPostAccountSuccess(t *testing.T) {
	app := testApp()
	res := post("/accounts", app.PostAccount, map[string]string{
		"username": "foo",
		"password": "bar",
	})

	assertCode(t, res, http.StatusCreated)
	assertSession(t, res)
	assertIdTokenResponse(t, res, app.Config)
}

func TestPostAccountSuccessWithSession(t *testing.T) {
	app := testApp()
	account_id := 8642
	session := createSession(app.RefreshTokenStore, app.Config, account_id)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(account_id)
	if err != nil {
		t.Fatal(err)
	}
	refreshToken := refreshTokens[0]

	post("/accounts", app.PostAccount, map[string]string{
		"username": "foo",
		"password": "bar",
	},
		func(req *http.Request) { req.AddCookie(session) },
	)

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

	var tests = []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"username", "MISSING"}, {"password", "MISSING"}}},
	}

	for _, tt := range tests {
		res := post("/accounts", app.PostAccount, map[string]string{
			"username": tt.username,
			"password": tt.password,
		})

		assertCode(t, res, http.StatusUnprocessableEntity)
		assertErrors(t, res, tt.errors)
	}
}
