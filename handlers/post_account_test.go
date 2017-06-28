package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	post("/accounts", app.PostAccount, map[string]string{
		"username": "foo",
		"password": "bar",
	},
		func(req *http.Request) *http.Request {
			req.AddCookie(session)
			return req
		},
	)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
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
