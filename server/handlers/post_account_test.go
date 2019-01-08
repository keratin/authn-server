package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAccountSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.PostForm("/accounts", url.Values{
		"username": []string{"foo"},
		"password": []string{"0a0b0c0"},
	})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertSession(t, app.Config, res.Cookies())
	test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
}

func TestPostAccountSuccessWithSession(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	accountID := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(accountID)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(session)
	_, err = client.PostForm("/accounts", url.Values{
		"username": []string{"foo"},
		"password": []string{"0a0b0c0"},
	})
	require.NoError(t, err)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostAccountFailure(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	var testCases = []struct {
		username string
		password string
		errors   services.FieldErrors
	}{
		{"", "", services.FieldErrors{{"username", "MISSING"}, {"password", "MISSING"}}},
	}

	for _, tc := range testCases {
		client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
		res, err := client.PostForm("/accounts", url.Values{
			"username": []string{tc.username},
			"password": []string{tc.password},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, tc.errors)
	}
}
