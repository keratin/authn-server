package handlers_test

import (
	"github.com/keratin/authn-server/app/models"
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostAccountsImport(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	formTests := []struct {
		Description string
		Values      url.Values
		Assertion   func(acc *models.Account, res *http.Response)
	}{
		{
			Description: "Should import someone (Form encoded)",
			Values:      url.Values{"username": []string{"someone@app.com"}, "password": []string{"secret"}},
			Assertion: func(acc *models.Account, res *http.Response) {
				test.AssertData(t, res, map[string]int{"id": acc.ID})
			},
		},
		{
			Description: "Should import a locked user (Form encoded)",
			Values:      url.Values{"username": []string{"locked@app.com"}, "password": []string{"secret"}, "locked": []string{"true"}},
			Assertion: func(acc *models.Account, res *http.Response) {
				assert.True(t, acc.Locked)
			},
		},
		{
			Description: "Should import an unlocked user (Form encoded)",
			Values:      url.Values{"username": []string{"unlocked@app.com"}, "password": []string{"secret"}, "locked": []string{"false"}},
			Assertion: func(acc *models.Account, res *http.Response) {
				assert.False(t, acc.Locked)
			},
		},
	}

	for _, test := range formTests {
		t.Run(test.Description, func(t *testing.T) {
			res, err := client.PostForm("/accounts/import", test.Values)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, res.StatusCode)
			account, err := app.AccountStore.FindByUsername(test.Values.Get("username"))
			require.NoError(t, err)
			test.Assertion(account, res)
		})
	}

	t.Run("importing an unlocked user using JSON", func(t *testing.T) {
		res, err := client.PostJSON("/accounts/import", "{\"username\":\"jsonunlocked@app.com\",\"password\": \"secret\",\"locked\":\"false\"}")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		account, err := app.AccountStore.FindByUsername("jsonunlocked@app.com")
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
