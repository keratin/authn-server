package accounts_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAccountSuccess(t *testing.T) {
	app := test.App()
	res := test.Post("/accounts", accounts.PostAccount(app), map[string]string{
		"username": "foo",
		"password": "bar",
	})

	test.AssertCode(t, res, http.StatusCreated)
	test.AssertSession(t, res)
	test.AssertIdTokenResponse(t, res, app.Config)
}

func TestPostAccountSuccessWithSession(t *testing.T) {
	app := test.App()
	account_id := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, account_id)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(account_id)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	test.Post("/accounts", accounts.PostAccount(app), map[string]string{
		"username": "foo",
		"password": "bar",
	}, test.WithSession(session))

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostAccountFailure(t *testing.T) {
	app := test.App()

	var tests = []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"username", "MISSING"}, {"password", "MISSING"}}},
	}

	for _, tt := range tests {
		res := test.Post("/accounts", accounts.PostAccount(app), map[string]string{
			"username": tt.username,
			"password": tt.password,
		})

		test.AssertCode(t, res, http.StatusUnprocessableEntity)
		test.AssertErrors(t, res, tt.errors)
	}
}
