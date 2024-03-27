package services_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/require"
)

func TestAccountOauthEnder(t *testing.T) {

	t.Run("require password reset for an account registered with oauth flow", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("requirepasswordreset@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(account.ID, "test", "TESTID", "email", "TOKEN")
		require.NoError(t, err)

		result, err := services.AccountOauthEnder(accountStore, account.ID, []string{"test"})
		require.NoError(t, err)

		updatedAccount, err := accountStore.Find(account.ID)
		require.NoError(t, err)

		require.Equal(t, updatedAccount.RequireNewPassword, true)
		require.Equal(t, result.RequirePasswordReset, true)
	})

	t.Run("delete non existing oauth accounts", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		result, err := services.AccountOauthEnder(accountStore, account.ID, []string{"test"})
		require.NoError(t, err)

		oAccount, err := accountStore.GetOauthAccounts(account.ID)
		require.NoError(t, err)

		require.Equal(t, result.RequirePasswordReset, false)
		require.Equal(t, len(oAccount), 0)
	})

	t.Run("delete account", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		time.Sleep(5 * time.Second)

		err = accountStore.AddOauthAccount(account.ID, "test", "TESTID", "email", "TOKEN")
		require.NoError(t, err)

		result, err := services.AccountOauthEnder(accountStore, account.ID, []string{"test"})
		require.NoError(t, err)

		oAccount, err := accountStore.GetOauthAccounts(account.ID)
		require.NoError(t, err)

		require.Equal(t, result.RequirePasswordReset, false)
		require.Equal(t, len(oAccount), 0)
	})

	t.Run("delete multiple accounts", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(account.ID, "test", "TESTID", "email", "TOKEN")
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(account.ID, "trial", "TESTID", "email", "TOKEN")
		require.NoError(t, err)

		result, err := services.AccountOauthEnder(accountStore, account.ID, []string{"test", "trial"})
		require.NoError(t, err)

		oAccount, err := accountStore.GetOauthAccounts(account.ID)
		require.NoError(t, err)

		require.Equal(t, result.RequirePasswordReset, true)
		require.Equal(t, len(oAccount), 0)
	})
}
