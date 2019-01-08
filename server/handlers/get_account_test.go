package handlers_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	app := test.App()
	server := test.Server(app)
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
		assertGetAccountResponse(t, res, account)
	})
}

func assertGetAccountResponse(t *testing.T, res *http.Response, acc *models.Account) {
	// check that the response contains the expected json
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	responseData := struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Locked   bool   `json:"locked"`
		Deleted  bool   `json:"deleted_at"`
	}{}
	err := test.ExtractResult(res, &responseData)
	assert.NoError(t, err)

	assert.Equal(t, acc.Username, responseData.Username)
	assert.Equal(t, acc.ID, responseData.ID)
	assert.Equal(t, false, responseData.Locked)
	assert.Equal(t, false, responseData.Deleted)
}
