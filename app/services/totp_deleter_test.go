package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOTPDeleter(t *testing.T) {
	accountStore := mock.NewAccountStore()

	account, err := accountStore.Create("test user", []byte("password"))
	require.NoError(t, err)

	t.Run("no account", func(t *testing.T) {
		deleteErr := services.TOTPDeleter(accountStore, 0)
		assert.Error(t, deleteErr)
	})
	t.Run("no secret", func(t *testing.T) {
		deleteErr := services.TOTPDeleter(accountStore, account.ID)
		assert.Error(t, deleteErr)
	})

	t.Run("secret", func(t *testing.T) {
		set, setErr := accountStore.SetTOTPSecret(account.ID, []byte("test"))
		assert.True(t, set)
		assert.NoError(t, setErr)

		deleteErr := services.TOTPDeleter(accountStore, account.ID)
		assert.NoError(t, deleteErr)
	})
}
