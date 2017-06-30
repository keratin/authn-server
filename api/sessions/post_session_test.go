package sessions_test

import (
	"net/http"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostSessionSuccess(t *testing.T) {
	app := test.App()
	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	app.AccountStore.Create("foo", b)

	res := test.Post("/session", sessions.PostSession(app), map[string]string{
		"username": "foo",
		"password": "bar",
	})

	test.AssertCode(t, res, http.StatusCreated)
	test.AssertSession(t, res)
	test.AssertIdTokenResponse(t, res, app.Config)
}

func TestPostSessionSuccessWithSession(t *testing.T) {
	app := test.App()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	app.AccountStore.Create("foo", b)

	account_id := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, account_id)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(account_id)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	test.Post("/session", sessions.PostSession(app), map[string]string{
		"username": "foo",
		"password": "bar",
	}, test.WithSession(session))

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostSessionFailure(t *testing.T) {
	app := test.App()

	var tests = []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"credentials", "FAILED"}}},
	}

	for _, tt := range tests {
		res := test.Post("/session", sessions.PostSession(app), map[string]string{
			"username": tt.username,
			"password": tt.password,
		})

		test.AssertCode(t, res, http.StatusUnprocessableEntity)
		test.AssertErrors(t, res, tt.errors)
	}
}
