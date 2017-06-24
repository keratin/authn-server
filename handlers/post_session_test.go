package handlers_test

import (
	"net/http"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/services"
)

func TestPostSessionSuccess(t *testing.T) {
	app := testApp()
	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	app.AccountStore.Create("foo", b)

	res := post("/session", app.PostSession, map[string]string{
		"username": "foo",
		"password": "bar",
	})

	assertCode(t, res, http.StatusCreated)
	assertSession(t, res)
	assertIdTokenResponse(t, res, app.Config)
}

func TestPostSessionSuccessWithSession(t *testing.T) {
	app := testApp()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	app.AccountStore.Create("foo", b)

	account_id := 8642
	session := createSession(app.RefreshTokenStore, app.Config, account_id)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(account_id)
	if err != nil {
		t.Fatal(err)
	}
	refreshToken := refreshTokens[0]

	post("/session", app.PostSession, map[string]string{
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

func TestPostSessionFailure(t *testing.T) {
	app := testApp()

	var tests = []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"credentials", "FAILED"}}},
	}

	for _, tt := range tests {
		res := post("/session", app.PostSession, map[string]string{
			"username": tt.username,
			"password": tt.password,
		})

		assertCode(t, res, http.StatusUnprocessableEntity)
		assertErrors(t, res, tt.errors)
	}
}
