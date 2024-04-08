package handlers_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
)

func TestDeleteAccountOauth(t *testing.T) {
	app := test.App()

	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("success", func(t *testing.T) {
		account, err := app.AccountStore.Create("deleted-social-account@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "DELETEDID", "email", "TOKEN")
		require.NoError(t, err)

		url := fmt.Sprintf("/accounts/%d/oauth/%s", account.ID, "test")
		res, err := client.Delete(url)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})

	t.Run("user does not exist", func(t *testing.T) {
		res, err := client.Delete("/accounts/9999/oauth/test")
		require.NoError(t, err)

		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}
