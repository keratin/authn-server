package services_test

import (
	"sort"
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/require"
)

func TestAccountGetter(t *testing.T) {

	t.Run("get non existing account", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := services.AccountGetter(accountStore, 9999)

		require.NotNil(t, err)
		require.Nil(t, account)
	})

	t.Run("returns empty map when no oauth accounts", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		acc, err := accountStore.Create("user@keratin.tech", []byte("password"))
		require.NoError(t, err)

		account, err := services.AccountGetter(accountStore, acc.ID)
		require.NoError(t, err)

		require.Equal(t, 0, len(account.OauthAccounts))
	})

	t.Run("returns oauth accounts for different providers", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		acc, err := accountStore.Create("user@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(acc.ID, "test", "ID1", "email1", "TOKEN1")
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(acc.ID, "trial", "ID2", "email2", "TOKEN2")
		require.NoError(t, err)

		account, err := services.AccountGetter(accountStore, acc.ID)
		require.NoError(t, err)

		oAccounts := account.OauthAccounts

		sort.Slice(oAccounts, func(i, j int) bool {
			return oAccounts[i].ProviderID < oAccounts[j].ProviderID
		})

		require.Equal(t, 2, len(oAccounts))
		require.Equal(t, "test", oAccounts[0].Provider)
		require.Equal(t, "ID1", oAccounts[0].ProviderID)
		require.Equal(t, "email1", oAccounts[0].Email)

		require.Equal(t, "trial", oAccounts[1].Provider)
		require.Equal(t, "ID2", oAccounts[1].ProviderID)
		require.Equal(t, "email2", oAccounts[1].Email)
	})
}
