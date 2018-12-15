package handlers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/services"

	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchAccount(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	t.Run("unknown account", func(t *testing.T) {
		res, err := client.Patch("/accounts/999999", url.Values{"username": []string{"irrelevant"}})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("existing account", func(t *testing.T) {
		account, err := app.AccountStore.Create("one@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v", account.ID), url.Values{"username": []string{"newname"}})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		account, err = app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.Equal(t, "newname", account.Username)
	})

	t.Run("bad username", func(t *testing.T) {
		account, err := app.AccountStore.Create("two@test.com", []byte("bar"))
		require.NoError(t, err)

		res, err := client.Patch(fmt.Sprintf("/accounts/%v", account.ID), url.Values{"username": []string{""}})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{"username", services.ErrMissing}})
	})
}
