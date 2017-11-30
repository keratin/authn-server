package accounts_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	app := test.App()
	server := test.Server(app, accounts.Routes(app))
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Get("/accounts/999999")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("invalid account", func(t *testing.T) {
		res, err := client.Get("/accounts/some_string")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("valid account", func(t *testing.T) {
		account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Get(fmt.Sprintf("/accounts/%v", account.ID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		test.AssertGetAccountResponse(t, res, account)
	})
}
