package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthtoken "github.com/keratin/authn-server/app/tokens/oauth"
	oauthlib "github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
)

func TestGetOauthReturn(t *testing.T) {
	// start a fake oauth provider
	providerServer := httptest.NewServer(test.ProviderApp())
	defer providerServer.Close()

	// configure a client for the fake oauth provider
	providerClient := oauthlib.NewTestProvider(providerServer)

	// configure and start the authn test server
	app := test.App()
	app.OauthProviders["test"] = *providerClient
	server := test.Server(app)
	defer server.Close()

	// configure a client for the authn test server
	nonce := "rand123"
	client := route.NewClient(server.URL).WithCookie(&http.Cookie{
		Name:  app.Config.OAuthCookieName,
		Value: nonce,
	})
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	token, err := oauthtoken.New(app.Config, nonce, "https://localhost:9999/return")
	require.NoError(t, err)
	state, err := token.Sign(app.Config.OAuthSigningKey)
	require.NoError(t, err)

	t.Run("sign up new identity with new email", func(t *testing.T) {
		res, err := client.Get("/oauth/test/return?code=something&state=" + state)
		require.NoError(t, err)
		if !test.AssertRedirect(t, res, "https://localhost:9999/return") {
			return
		}
		test.AssertSession(t, app.Config, res.Cookies(), "oauth:test")

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

		res, err := client.WithCookie(session).Get("/oauth/test/return?code=existing@keratin.tech&state=" + state)
		require.NoError(t, err)
		if test.AssertRedirect(t, res, "https://localhost:9999/return") {
			test.AssertSession(t, app.Config, res.Cookies(), "oauth:test")
		}
	})

	t.Run("not connect new identity with current session that is already linked", func(t *testing.T) {
		account, err := app.AccountStore.Create("linked@keratin.tech", []byte("password"))
		require.NoError(t, err)
		err = app.AccountStore.AddOauthAccount(account.ID, "test", "PREVIOUSID", "email", "TOKEN")
		require.NoError(t, err)
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Get("/oauth/test/return?code=linked+alias@keratin.tech&state=" + state)
		require.NoError(t, err)
		test.AssertRedirect(t, res, "https://localhost:9999/return?status=failed")
	})

	t.Run("not connect provider account already linked", func(t *testing.T) {
		linkedAccount, err := app.AccountStore.Create("linked.account@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(linkedAccount.ID, "test", "LINKEDID", "email", "TOKEN")
		require.NoError(t, err)

		account, err := app.AccountStore.Create("registered.account@keratin.tech", []byte("password"))
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Get("/oauth/test/return?code=LINKEDID&state=" + state)
		require.NoError(t, err)
		test.AssertRedirect(t, res, "https://localhost:9999/return?status=failed")
	})

	t.Run("log in to existing identity", func(t *testing.T) {
		account, err := app.AccountStore.Create("registered@keratin.tech", []byte("password"))
		require.NoError(t, err)
		err = app.AccountStore.AddOauthAccount(account.ID, "test", "REGISTEREDID", "email", "TOKEN")
		require.NoError(t, err)

		// codes don't normally specify the id, but our test provider is set up to reflect the code
		// back as id and email.
		res, err := client.Get("/oauth/test/return?code=REGISTEREDID&state=" + state)
		require.NoError(t, err)
		if test.AssertRedirect(t, res, "https://localhost:9999/return") {
			test.AssertSession(t, app.Config, res.Cookies(), "oauth:test")
		}
	})

	t.Run("log in to locked identity", func(t *testing.T) {
		account, err := app.AccountStore.Create("locked@keratin.tech", []byte("password"))
		require.NoError(t, err)
		_, err = app.AccountStore.Lock(account.ID)
		require.NoError(t, err)

		res, err := client.Get("/oauth/test/return?code=locked@keratin.tech&state=" + state)
		require.NoError(t, err)
		test.AssertRedirect(t, res, "https://localhost:9999/return?status=failed")
	})

	t.Run("email collision", func(t *testing.T) {
		_, err := app.AccountStore.Create("collision@keratin.tech", []byte("password"))
		require.NoError(t, err)

		res, err := client.Get("/oauth/test/return?code=collision@keratin.tech&state=" + state)
		require.NoError(t, err)
		test.AssertRedirect(t, res, "https://localhost:9999/return?status=failed")
	})

	t.Run("without nonce cookie", func(t *testing.T) {
		client := route.NewClient(server.URL)
		res, err := client.Get("/oauth/test/return?code=something&state=" + state)
		require.NoError(t, err)
		test.AssertRedirect(t, res, "http://test.com")
	})

	t.Run("with tampered state", func(t *testing.T) {
		res, err := client.Get("/oauth/test/return?code=something&state=TAMPERED")
		require.NoError(t, err)
		test.AssertRedirect(t, res, "http://test.com")
	})
}
