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

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	expired, err := accountStore.Create("expired@keratin.tech", []byte("old"))
	require.NoError(t, err)
	err = accountStore.RequireNewPassword(expired.Id)
	require.NoError(t, err)

	lockedAccount, err := accountStore.Create("locked@keratin.tech", []byte("old"))
	require.NoError(t, err)
	err = accountStore.Lock(lockedAccount.Id)
	require.NoError(t, err)

	t.Run("it resets RequireNoPassword", func(t *testing.T) {
		err := services.PasswordChanger(accountStore, cfg, expired.Id, "0a0b0c0d0e0f")
		assert.NoError(t, err)
		account, err := accountStore.Find(expired.Id)
		require.NoError(t, err)
		assert.False(t, account.RequireNewPassword)
		assert.NotEqual(t, expired.Password, account.Password)
	})

	failureCases := []struct {
		account_id int
		password   string
		errors     services.FieldErrors
	}{
		{0, "0a0b0c0d0e0f", services.FieldErrors{{"account", "NOT_FOUND"}}},
		{account.Id, "abc", services.FieldErrors{{"password", "INSECURE"}}},
		{account.Id, "", services.FieldErrors{{"password", "MISSING"}}},
		{lockedAccount.Id, "0a0b0c0d0e0f", services.FieldErrors{{"account", "LOCKED"}}},
	}

	for _, fc := range failureCases {
		t.Run(fc.errors.Error(), func(t *testing.T) {
			err := services.PasswordChanger(accountStore, cfg, fc.account_id, fc.password)
			assert.Equal(t, fc.errors, err)
		})
	}
}
