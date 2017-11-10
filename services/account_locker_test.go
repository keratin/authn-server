package services_test

import (
	"testing"

	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountLocker(t *testing.T) {
	store := mock.NewAccountStore()
	refreshStore := mock.NewRefreshTokenStore()

	t.Run("logged in account", func(t *testing.T) {
		account, err := store.Create("loggedin@keratin.tech", []byte("password"))
		require.NoError(t, err)
		token1, err := refreshStore.Create(account.ID)
		require.NoError(t, err)

		errs := services.AccountLocker(store, refreshStore, account.ID)
		assert.Empty(t, errs)

		id, err := refreshStore.Find(token1)
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("locked account", func(t *testing.T) {
		account, err := store.Create("locked@keratin.tech", []byte("password"))
		require.NoError(t, err)
		err = store.Lock(account.ID)
		require.NoError(t, err)

		errs := services.AccountLocker(store, refreshStore, account.ID)
		assert.Empty(t, errs)

		acct, err := store.Find(account.ID)
		require.NoError(t, err)
		assert.True(t, acct.Locked)
	})

	t.Run("unlocked account", func(t *testing.T) {
		account, err := store.Create("unlocked@keratin.tech", []byte("password"))
		require.NoError(t, err)

		errs := services.AccountLocker(store, refreshStore, account.ID)
		assert.Empty(t, errs)

		acct, err := store.Find(account.ID)
		require.NoError(t, err)
		assert.True(t, acct.Locked)
	})

	t.Run("unknown account", func(t *testing.T) {
		errs := services.AccountLocker(store, refreshStore, 123456789)
		assert.Equal(t, services.FieldErrors{{"account", services.ErrNotFound}}, errs)
	})
}
