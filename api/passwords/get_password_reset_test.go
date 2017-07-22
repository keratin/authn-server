package passwords_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/api/passwords"
	"github.com/keratin/authn-server/api/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPasswordReset(t *testing.T) {
	app := test.App()
	server := test.Server(app, passwords.Routes(app))
	defer server.Close()

	client := test.NewClient(server).Referred(app.Config)

	t.Run("known account", func(t *testing.T) {
		_, err := app.AccountStore.Create("known@keratin.tech", []byte("pwd"))
		require.NoError(t, err)

		res, err := client.Get("/password/reset?username=known@keratin.tech")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// TODO: assert go routine?
	})

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Get("/password/reset?username=unknown@keratin.tech")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
