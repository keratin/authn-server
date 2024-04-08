package handlers_test

import (
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

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
	oAccounts := []map[string]interface{}{}

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

	for _, oAcc := range oAccs {
		oAccounts = append(oAccounts, map[string]interface{}{
			"provider":            oAcc.Provider,
			"provider_account_id": oAcc.ProviderID,
			"email":               oAcc.Email,
		})
	}

	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	assert.Equal(t, responseData, response{
		ID:                acc.ID,
		Username:          acc.Username,
		OauthAccounts:     oAccounts,
		LastLoginAt:       "",
		PasswordChangedAt: acc.PasswordChangedAt.Format(time.RFC3339),
		Locked:            false,
		Deleted:           false,
	})
}
