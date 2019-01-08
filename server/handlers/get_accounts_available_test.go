package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccountsAvailable(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	account, err := app.AccountStore.Create("existing@test.com", []byte("bar"))
	require.NoError(t, err)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])

	t.Run("known username", func(t *testing.T) {
		res, err := client.Get("/accounts/available?username=" + account.Username)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	})

	t.Run("unknown username", func(t *testing.T) {
		res, err := client.Get("/accounts/available?username=unknown@test.com")
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
