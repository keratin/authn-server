package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(&http.Cookie{
		Name:  app.Config.OAuthCookieName,
		Value: "",
	})

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("unauthorized", func(t *testing.T) {
		res, err := client.Delete("/oauth/test")
		require.NoError(t, err)

		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})

	t.Run("success", func(t *testing.T) {
		account, err := app.AccountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "DELETEDID", "email", "TOKEN")
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Delete("/oauth/test")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})
}
