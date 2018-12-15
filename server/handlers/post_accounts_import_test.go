package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAccountsImport(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	t.Run("importing someone", func(t *testing.T) {
		res, err := client.PostForm("/accounts/import", url.Values{
			"username": []string{"someone@app.com"},
			"password": []string{"secret"},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		account, err := app.AccountStore.FindByUsername("someone@app.com")
		require.NoError(t, err)
		test.AssertData(t, res, map[string]int{"id": account.ID})
	})

	t.Run("importing a locked user", func(t *testing.T) {
		res, err := client.PostForm("/accounts/import", url.Values{
			"username": []string{"locked@app.com"},
			"password": []string{"secret"},
			"locked":   []string{"true"},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		account, err := app.AccountStore.FindByUsername("locked@app.com")
		require.NoError(t, err)
		assert.True(t, account.Locked)
	})

	t.Run("importing an unlocked user", func(t *testing.T) {
		res, err := client.PostForm("/accounts/import", url.Values{
			"username": []string{"unlocked@app.com"},
			"password": []string{"secret"},
			"locked":   []string{"false"},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		account, err := app.AccountStore.FindByUsername("someone@app.com")
		require.NoError(t, err)
		assert.False(t, account.Locked)
	})

	t.Run("importing an invalid user", func(t *testing.T) {
		res, err := client.PostForm("/accounts/import", url.Values{
			"username": []string{"invalid@app.com"},
			"password": []string{""},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{"password", "MISSING"}})
	})

}
