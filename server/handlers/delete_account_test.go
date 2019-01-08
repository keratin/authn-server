package handlers_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteAccount(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Delete("/accounts/999999")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("unarchived account", func(t *testing.T) {
		account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Delete(fmt.Sprintf("/accounts/%v", account.ID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, account.DeletedAt)
	})

	t.Run("archived account", func(t *testing.T) {
		account, err := app.AccountStore.Create("locked@test.com", []byte("bar"))
		require.NoError(t, err)
		app.AccountStore.Archive(account.ID)

		res, err := client.Delete(fmt.Sprintf("/accounts/%v", account.ID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, account.DeletedAt)
	})
}
