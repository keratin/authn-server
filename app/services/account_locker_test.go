package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountLocker(t *testing.T) {
	accountStore := mock.NewAccountStore()
	refreshStore := mock.NewRefreshTokenStore()

	t.Run("logged in account", func(t *testing.T) {
		account, err := accountStore.Create("loggedin@keratin.tech", []byte("password"))
		require.NoError(t, err)
		token1, err := refreshStore.Create(account.ID)
		require.NoError(t, err)

		errs := services.AccountLocker(accountStore, refreshStore, account.ID)
		assert.Empty(t, errs)

		id, err := refreshStore.Find(token1)
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("locked account", func(t *testing.T) {
		account, err := accountStore.Create("locked@keratin.tech", []byte("password"))
		require.NoError(t, err)
		_, err = accountStore.Lock(account.ID)
		require.NoError(t, err)

		errs := services.AccountLocker(accountStore, refreshStore, account.ID)
		assert.Empty(t, errs)

		acct, err := accountStore.Find(account.ID)
		require.NoError(t, err)
		assert.True(t, acct.Locked)
	})

	t.Run("unlocked account", func(t *testing.T) {
		account, err := accountStore.Create("unlocked@keratin.tech", []byte("password"))
		require.NoError(t, err)

		errs := services.AccountLocker(accountStore, refreshStore, account.ID)
		assert.Empty(t, errs)

		acct, err := accountStore.Find(account.ID)
		require.NoError(t, err)
		assert.True(t, acct.Locked)
	})

	t.Run("unknown account", func(t *testing.T) {
		errs := services.AccountLocker(accountStore, refreshStore, 123456789)
		assert.Equal(t, services.FieldErrors{{"account", services.ErrNotFound}}, errs)
	})
}
