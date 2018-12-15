package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountArchiver(t *testing.T) {
	accountStore := mock.NewAccountStore()
	refreshStore := mock.NewRefreshTokenStore()

	t.Run("existing account", func(t *testing.T) {
		account, err := accountStore.Create("test@keratin.tech", []byte("password"))
		require.NoError(t, err)

		errs := services.AccountArchiver(accountStore, refreshStore, account.ID)
		assert.Empty(t, errs)

		acct, err := accountStore.Find(account.ID)
		require.NoError(t, err)
		assert.Empty(t, acct.Username)
		assert.Empty(t, acct.Password)
		assert.NotEmpty(t, acct.DeletedAt)
	})

	t.Run("logged in account", func(t *testing.T) {
		account, err := accountStore.Create("loggedin@keratin.tech", []byte("password"))
		require.NoError(t, err)
		token1, err := refreshStore.Create(account.ID)
		require.NoError(t, err)

		errs := services.AccountArchiver(accountStore, refreshStore, account.ID)
		assert.Empty(t, errs)

		id, err := refreshStore.Find(token1)
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("unknown account", func(t *testing.T) {
		errs := services.AccountArchiver(accountStore, refreshStore, 123456789)
		assert.Equal(t, services.FieldErrors{{"account", services.ErrNotFound}}, errs)
	})
}
