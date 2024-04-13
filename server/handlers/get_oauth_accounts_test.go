package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
)

func TestGetOauthInfo(t *testing.T) {
	app := test.App()

	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(&http.Cookie{
		Name:  app.Config.OAuthCookieName,
		Value: "",
	})

	t.Run("unauthorized", func(t *testing.T) {
		res, err := client.Get("/oauth/accounts")
		require.NoError(t, err)

		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})

	t.Run("success", func(t *testing.T) {
		var expected struct {
			Result []struct {
				Email             string `json:"email"`
				Provider          string `json:"provider"`
				ProviderAccountID string `json:"provider_account_id"`
			}
		}

		account, err := app.AccountStore.Create("get-oauth-info@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "ID", "email", "TOKEN")
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Get("/oauth/accounts")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		err = json.Unmarshal(test.ReadBody(res), &expected)
		require.NoError(t, err)

		require.Equal(t, len(expected.Result), 1)
		require.Equal(t, expected.Result[0].Email, "email")
		require.Equal(t, expected.Result[0].Provider, "test")
		require.Equal(t, expected.Result[0].ProviderAccountID, "ID")
	})
}
