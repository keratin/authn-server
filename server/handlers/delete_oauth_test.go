package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthlib "github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
)

func TestDeleteOauthAccount(t *testing.T) {
	providerServer := httptest.NewServer(test.ProviderApp())
	defer providerServer.Close()

	app := test.App()
	app.OauthProviders["test"] = *oauthlib.NewTestProvider(providerServer)

	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).WithCookie(&http.Cookie{
		Name:  app.Config.OAuthCookieName,
		Value: "",
	})

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("unanthorized", func(t *testing.T) {
		res, err := client.Delete("/oauth/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("delete social account", func(t *testing.T) {
		account, err := app.AccountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "DELETEDID", "TOKEN")
		require.NoError(t, err)

		time.Sleep(5 * time.Second)

		_, err = app.AccountStore.SetPassword(account.ID, []byte("password"))
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Delete("/oauth/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})
}
