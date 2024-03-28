package handlers_test

import (
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
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

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "ID1", "email", "TOKEN1")
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "trial", "ID2", "email", "TOKEN2")
		require.NoError(t, err)

		oauthAccounts, err := app.AccountStore.GetOauthAccounts(account.ID)
		require.NoError(t, err)

		res, err := client.Get(fmt.Sprintf("/accounts/%v", account.ID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assertGetAccountResponse(t, res, account, oauthAccounts)
	})
}

func assertGetAccountResponse(t *testing.T, res *http.Response, acc *models.Account, oAccs []*models.OauthAccount) {
	// check that the response contains the expected json
	type response struct {
		ID                int                      `json:"id"`
		Username          string                   `json:"username"`
		OauthAccounts     []map[string]interface{} `json:"oauth_accounts"`
		LastLoginAt       string                   `json:"last_login_at"`
		PasswordChangedAt string                   `json:"password_changed_at"`
		Locked            bool                     `json:"locked"`
		Deleted           bool                     `json:"deleted_at"`
	}

	var responseData response
	err := test.ExtractResult(res, &responseData)
	assert.NoError(t, err)

	sort.Slice(oAccs, func(i, j int) bool {
		return oAccs[i].Provider < oAccs[j].Provider
	})

	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	assert.Equal(t, responseData, response{
		ID:       acc.ID,
		Username: acc.Username,
		OauthAccounts: []map[string]interface{}{
			{
				"provider":            "test",
				"provider_account_id": oAccs[0].ProviderID,
				"email":               oAccs[0].Email,
			},
			{
				"provider":            "trial",
				"provider_account_id": oAccs[1].ProviderID,
				"email":               oAccs[1].Email,
			},
		},
		LastLoginAt:       "",
		PasswordChangedAt: acc.PasswordChangedAt.Format("2006-01-02T15:04:05Z07:00"),
		Locked:            false,
		Deleted:           false,
	})
}
