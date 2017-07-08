package accounts_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchAccountLock(t *testing.T) {
	app := test.App()
	server := test.Server(app, accounts.Routes(app))
	defer server.Close()

	client := test.NewClient(server).Authenticated(app.Config)

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Patch("/accounts/999999/lock", url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("unlocked account", func(t *testing.T) {
		account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v/lock", account.Id), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.Id)
		require.NoError(t, err)
		assert.True(t, account.Locked)
	})

	t.Run("locked account", func(t *testing.T) {
		account, err := app.AccountStore.Create("locked@test.com", []byte("bar"))
		require.NoError(t, err)
		app.AccountStore.Lock(account.Id)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v/lock", account.Id), url.Values{})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.Id)
		require.NoError(t, err)
		assert.True(t, account.Locked)
	})
}
