package accounts_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAccountSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app, accounts.Routes(app))
	defer server.Close()

	client := test.NewClient(server).Referred(app.Config)
	res, err := client.PostForm("/accounts", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
	})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertSession(t, app.Config, res.Cookies())
	test.AssertIdTokenResponse(t, res, app.Config)
}

func TestPostAccountSuccessWithSession(t *testing.T) {
	app := test.App()
	server := test.Server(app, accounts.Routes(app))
	defer server.Close()

	account_id := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, account_id)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(account_id)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	client := test.NewClient(server).Referred(app.Config).WithSession(session)
	_, err = client.PostForm("/accounts", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
	})
	require.NoError(t, err)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostAccountFailure(t *testing.T) {
	app := test.App()
	server := test.Server(app, accounts.Routes(app))
	defer server.Close()

	var testCases = []struct {
		username string
		password string
		errors   []services.Error
	}{
		{"", "", []services.Error{{"username", "MISSING"}, {"password", "MISSING"}}},
	}

	for _, tc := range testCases {
		client := test.NewClient(server).Referred(app.Config)
		res, err := client.PostForm("/accounts", url.Values{
			"username": []string{tc.username},
			"password": []string{tc.password},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, tc.errors)
	}
}
