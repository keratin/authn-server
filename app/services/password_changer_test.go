package services_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordChanger(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
		BcryptCost:            4,
		PasswordMinComplexity: 1,
	}

	invoke := func(id int, currentPassword string, password string) error {
		return services.PasswordChanger(accountStore, &ops.LogReporter{FieldLogger: logrus.New()}, cfg, id, currentPassword, password)
	}

	factory := func(username string, password string) (*models.Account, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
		if err != nil {
			return nil, errors.Wrap(err, "bcrypt")
		}

		return accountStore.Create(username, hash)
	}

	account, err := factory("existing@keratin.tech", "old")
	require.NoError(t, err)

	t.Run("it resets RequireNoPassword", func(t *testing.T) {
		expired, err := factory("expired@keratin.tech", "old")
		require.NoError(t, err)
		_, err = accountStore.RequireNewPassword(expired.ID)
		require.NoError(t, err)

		err = invoke(expired.ID, "old", "0a0b0c0d0e0f")
		assert.NoError(t, err)

		account, err := accountStore.Find(expired.ID)
		require.NoError(t, err)
		assert.False(t, account.RequireNewPassword)
		assert.NotEqual(t, expired.Password, account.Password)
	})

	t.Run("with an unknown account", func(t *testing.T) {
		err := invoke(0, "unknown", "0ab0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "NOT_FOUND"}}, err)
	})

	t.Run("with a locked account", func(t *testing.T) {
		lockedAccount, err := factory("locked@keratin.tech", "old")
		require.NoError(t, err)
		_, err = accountStore.Lock(lockedAccount.ID)
		require.NoError(t, err)

		err = invoke(lockedAccount.ID, "old", "0ab0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("with an insecure password", func(t *testing.T) {
		err := invoke(account.ID, "old", "abc")
		assert.Equal(t, services.FieldErrors{{"password", "INSECURE"}}, err)
	})

	t.Run("with a missing password", func(t *testing.T) {
		err := invoke(account.ID, "old", "")
		assert.Equal(t, services.FieldErrors{{"password", "MISSING"}}, err)
	})

	t.Run("with the wrong current password", func(t *testing.T) {
		err := invoke(account.ID, "wrong", "")
		assert.Equal(t, services.FieldErrors{{"credentials", "FAILED"}}, err)
	})
}
