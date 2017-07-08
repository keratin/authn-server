package accounts_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccountsAvailable(t *testing.T) {
	app := test.App()
	server := test.Server(app, accounts.Routes(app))
	defer server.Close()

	account, err := app.AccountStore.Create("existing@test.com", []byte("bar"))
	require.NoError(t, err)

	client := test.NewClient(server).Referred(app.Config)

	t.Run("known username", func(t *testing.T) {
		res, err := client.Get("/accounts/available?username=" + account.Username)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("unknown username", func(t *testing.T) {
		res, err := client.Get("/accounts/available?username=unknown@test.com")
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	})
}
