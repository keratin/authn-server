package services_test

import (
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordSetter(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &config.Config{
		BcryptCost:            4,
		PasswordMinComplexity: 1,
	}

	invoke := func(id int, password string) error {
		return services.PasswordSetter(accountStore, cfg, id, password)
	}

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	t.Run("sets password", func(t *testing.T) {
		err = accountStore.RequireNewPassword(account.Id)
		require.NoError(t, err)

		err := invoke(account.Id, "0a0b0c0d0e0f0")
		assert.NoError(t, err)

		after, err := accountStore.Find(account.Id)
		require.NoError(t, err)
		assert.NotEqual(t, account.Password, after.Password)
		assert.False(t, account.RequireNewPassword)
	})

	t.Run("missing password", func(t *testing.T) {
		err := invoke(account.Id, "")
		assert.Equal(t, services.FieldErrors{{"password", "MISSING"}}, err)
	})

	t.Run("insecure password", func(t *testing.T) {
		err := invoke(account.Id, "abc")
		assert.Equal(t, services.FieldErrors{{"password", "INSECURE"}}, err)
	})
}
