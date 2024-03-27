package services_test

import (
	"sort"
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/require"
)

func TestAccountOauthGetter(t *testing.T) {
	t.Run("returns empty map when no oauth accounts", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("user@keratin.tech", []byte("password"))
		require.NoError(t, err)

		accountOauth, err := services.AccountOauthGetter(accountStore, account.ID)
		require.NoError(t, err)

		require.Equal(t, 0, len(accountOauth))
	})

	t.Run("returns oauth accounts for different providers", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("user@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(account.ID, "test", "ID1", "TOKEN1")
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(account.ID, "trial", "ID2", "TOKEN2")
		require.NoError(t, err)

		accountOauth, err := services.AccountOauthGetter(accountStore, 1)
		require.NoError(t, err)

		sort.Slice(accountOauth, func(i, j int) bool {
			return accountOauth[i]["provider_id"].(string) < accountOauth[j]["provider_id"].(string)
		})

		require.Equal(
			t,
			[]map[string]interface{}{
				{
					"provider":    "test",
					"provider_id": "ID1",
				},
				{
					"provider":    "trial",
					"provider_id": "ID2",
				},
			},
			accountOauth,
		)
	})
}
