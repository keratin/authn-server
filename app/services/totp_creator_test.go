package services_test

import (
	"errors"
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOTPCreator(t *testing.T) {
	accountStore := mock.NewAccountStore()
	totpCache := mock.NewTOTPCache(0)

	account, err := accountStore.Create("test user", []byte("password"))
	require.NoError(t, err)

	audience := &route.Domain{
		Hostname: "testhost",
		Port:     "80",
	}

	t.Run("no account", func(t *testing.T) {
		key, createErr := services.TOTPCreator(accountStore, totpCache, 0, audience)
		assert.Nil(t, key)
		assert.Error(t, createErr)
	})

	t.Run("account exists", func(t *testing.T) {
		key, createErr := services.TOTPCreator(accountStore, totpCache, account.ID, audience)
		require.NoError(t, createErr)
		require.NotNil(t, key)
		gotKey, gotErr := totpCache.LoadTOTPSecret(account.ID)
		assert.Nil(t, gotErr)
		assert.Equal(t, string(gotKey), key.Secret())

		t.Run("already enrolled", func(t *testing.T) {
			set, setErr := accountStore.SetTOTPSecret(account.ID, []byte(key.Secret()))
			require.True(t, set)
			require.NoError(t, setErr)

			key, createErr = services.TOTPCreator(accountStore, totpCache, account.ID, audience)
			assert.Nil(t, key)
			assert.Error(t, createErr)
			assert.True(t, errors.Is(createErr, services.ErrExistingTOTPSecret))
		})
	})

	t.Run("account exists - cache error", func(t *testing.T) {
		key, createErr := services.TOTPCreator(accountStore, mock.NewTOTPCache(account.ID), account.ID, audience)
		assert.Error(t, createErr)
		assert.Nil(t, key)
	})
}
