package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordExpirer(t *testing.T) {
	accountStore := mock.NewAccountStore()
	refreshStore := mock.NewRefreshTokenStore()

	t.Run("active account", func(t *testing.T) {
		account, err := accountStore.Create("active", []byte("secret"))
		require.NoError(t, err)
		token1, err := refreshStore.Create(account.ID)
		require.NoError(t, err)
		token2, err := refreshStore.Create(account.ID)
		require.NoError(t, err)

		errors := services.PasswordExpirer(accountStore, refreshStore, account.ID)
		assert.Empty(t, errors)

		account, err = accountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, account.RequireNewPassword)

		id, err := refreshStore.Find(token1)
		require.NoError(t, err)
		assert.Empty(t, id)
		id, err = refreshStore.Find(token2)
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("unknown account", func(t *testing.T) {
		errors := services.PasswordExpirer(accountStore, refreshStore, 0)
		assert.Equal(t, services.FieldErrors{{"account", services.ErrNotFound}}, errors)
	})
}
