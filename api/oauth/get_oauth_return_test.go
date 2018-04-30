package oauth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/keratin/authn-server/api/oauth"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/lib/route"
)

func TestGetOauthReturn(t *testing.T) {
	// start a fake oauth provider
	providerServer := httptest.NewServer(http.HandlerFunc(test.ProviderApp))
	defer providerServer.Close()

	// configure a client for the fake oauth provider
	providerClient := test.NewOauthProvider(providerServer)

	// configure and start the authn test server
	app := test.App()
	app.OauthProviders["test"] = providerClient
	server := test.Server(app, oauth.Routes(app))
	defer server.Close()

	// configure a client for the authn test server
	client := route.NewClient(server.URL)
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("sign up new identity with new email", func(t *testing.T) {
		res, err := client.Get("/oauth/test/return?code=something")
		require.NoError(t, err)
		if test.AssertRedirect(t, res, "http://localhost:9999/TODO/SUCCESS") {
			test.AssertSession(t, app.Config, res.Cookies())
		}

		// creates an account
		account, err := app.AccountStore.FindByOauthAccount("test", "something")
		require.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, "something", account.Username)
	})

	t.Run("connect new identity with current session", func(t *testing.T) {
		account, err := app.AccountStore.Create("existing@keratin.tech", []byte("password"))
		require.NoError(t, err)
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)
		res, err := client.WithCookie(session).Get("/oauth/test/return?code=existing@keratin.tech")
		require.NoError(t, err)
		if test.AssertRedirect(t, res, "http://localhost:9999/TODO/SUCCESS") {
			test.AssertSession(t, app.Config, res.Cookies())
		}
	})

	t.Run("log in to existing identity", func(t *testing.T) {
		account, err := app.AccountStore.Create("registered@keratin.tech", []byte("password"))
		require.NoError(t, err)
		err = app.AccountStore.AddOauthAccount(account.ID, "test", "REGISTEREDID", "TOKEN")
		require.NoError(t, err)

		// codes don't normally specify the id, but our test provider is set up to reflect the code
		// back as id and email.
		res, err := client.Get("/oauth/test/return?code=REGISTEREDID")
		require.NoError(t, err)
		if test.AssertRedirect(t, res, "http://localhost:9999/TODO/SUCCESS") {
			test.AssertSession(t, app.Config, res.Cookies())
		}
	})

	t.Run("email collision", func(t *testing.T) {
		_, err := app.AccountStore.Create("collision@keratin.tech", []byte("password"))
		require.NoError(t, err)
		res, err := client.Get("/oauth/test/return?code=collision@keratin.tech")
		require.NoError(t, err)
		test.AssertRedirect(t, res, "http://localhost:9999/TODO/FAILURE")
	})

	t.Run("connect new identity with current session that is already linked", func(t *testing.T) {
		account, err := app.AccountStore.Create("linked@keratin.tech", []byte("password"))
		require.NoError(t, err)
		app.AccountStore.AddOauthAccount(account.ID, "test", "PREVIOUSID", "TOKEN")
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)
		res, err := client.WithCookie(session).Get("/oauth/test/return?code=linked@keratin.tech")
		require.NoError(t, err)
		test.AssertRedirect(t, res, "http://localhost:9999/TODO/FAILURE")
	})
}
