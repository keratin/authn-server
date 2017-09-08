package services_test

import (
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordChanger(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &config.Config{
		BcryptCost:            4,
		PasswordMinComplexity: 1,
	}

	invoke := func(id int, password string) error {
		return services.PasswordChanger(accountStore, cfg, id, password)
	}

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	t.Run("it resets RequireNoPassword", func(t *testing.T) {
		expired, err := accountStore.Create("expired@keratin.tech", []byte("old"))
		require.NoError(t, err)
		err = accountStore.RequireNewPassword(expired.ID)
		require.NoError(t, err)

		err = invoke(expired.ID, "0a0b0c0d0e0f")
		assert.NoError(t, err)

		account, err := accountStore.Find(expired.ID)
		require.NoError(t, err)
		assert.False(t, account.RequireNewPassword)
		assert.NotEqual(t, expired.Password, account.Password)
	})

	t.Run("with an unknown account", func(t *testing.T) {
		err := invoke(0, "0ab0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "NOT_FOUND"}}, err)
	})

	t.Run("with a locked account", func(t *testing.T) {
		lockedAccount, err := accountStore.Create("locked@keratin.tech", []byte("old"))
		require.NoError(t, err)
		err = accountStore.Lock(lockedAccount.ID)
		require.NoError(t, err)

		err = invoke(lockedAccount.ID, "0ab0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("with an insecure password", func(t *testing.T) {
		err := invoke(account.ID, "abc")
		assert.Equal(t, services.FieldErrors{{"password", "INSECURE"}}, err)
	})

	t.Run("with a missing password", func(t *testing.T) {
		err := invoke(account.ID, "")
		assert.Equal(t, services.FieldErrors{{"password", "MISSING"}}, err)
	})
}
