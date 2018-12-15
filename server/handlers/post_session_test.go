package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostSessionSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	app.AccountStore.Create("foo", b)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.PostForm("/session", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
	})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertSession(t, app.Config, res.Cookies())
	test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
}

func TestPostSessionSuccessWithSession(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	app.AccountStore.Create("foo", b)

	accountID := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(accountID)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(session)
	_, err = client.PostForm("/session", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
	})
	require.NoError(t, err)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostSessionFailure(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	var testCases = []struct {
		username string
		password string
		errors   services.FieldErrors
	}{
		{"", "", services.FieldErrors{{"credentials", "FAILED"}}},
	}

	for _, tc := range testCases {
		client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
		res, err := client.PostForm("/session", url.Values{
			"username": []string{tc.username},
			"password": []string{tc.password},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, tc.errors)
	}
}
