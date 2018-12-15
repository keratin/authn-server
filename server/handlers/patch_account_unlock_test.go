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

func TestPatchAccountUnlock(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Patch("/accounts/999999/unlock", url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("unlocked account", func(t *testing.T) {
		account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v/unlock", account.ID), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.False(t, account.Locked)
	})

	t.Run("locked account", func(t *testing.T) {
		account, err := app.AccountStore.Create("locked@test.com", []byte("bar"))
		require.NoError(t, err)
		app.AccountStore.Lock(account.ID)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v/unlock", account.ID), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.False(t, account.Locked)
	})
}
