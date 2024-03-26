package services_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountOauthEnder(t *testing.T) {

	t.Run("require password reset before delete an account register with oauth flow", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("requirepasswordreset@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = accountStore.AddOauthAccount(account.ID, "test", "TESTID", "TOKEN")
		require.NoError(t, err)

		err = services.AccountOauthEnder(accountStore, account.ID, "test")
		assert.Equal(t, err, services.FieldErrors{{Field: "password", Message: services.ErrPasswordResetRequired}})
	})

	t.Run("delete account", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		account, err := accountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		time.Sleep(5 * time.Second)

		err = accountStore.AddOauthAccount(account.ID, "test", "TESTID", "TOKEN")
		require.NoError(t, err)

		err = services.AccountOauthEnder(accountStore, account.ID, "test")
		require.NoError(t, err)
	})
}
