package handlers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchAccountExpirePassword(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Patch("/accounts/999999/expire_password", url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("active account", func(t *testing.T) {
		account, err := app.AccountStore.Create("active@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v/expire_password", account.ID), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.True(t, account.RequireNewPassword)
	})
}
