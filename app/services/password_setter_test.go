package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/ops"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordSetter(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
		BcryptCost:            4,
		PasswordMinComplexity: 1,
	}

	invoke := func(id int, password string) error {
		return services.PasswordSetter(accountStore, &ops.LogReporter{logrus.New()}, cfg, id, password)
	}

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	t.Run("sets password", func(t *testing.T) {
		_, err = accountStore.RequireNewPassword(account.ID)
		require.NoError(t, err)

		err := invoke(account.ID, "0a0b0c0d0e0f0")
		assert.NoError(t, err)

		after, err := accountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEqual(t, account.Password, after.Password)
		assert.False(t, account.RequireNewPassword)
	})

	t.Run("missing password", func(t *testing.T) {
		err := invoke(account.ID, "")
		assert.Equal(t, services.FieldErrors{{"password", "MISSING"}}, err)
	})

	t.Run("insecure password", func(t *testing.T) {
		err := invoke(account.ID, "abc")
		assert.Equal(t, services.FieldErrors{{"password", "INSECURE"}}, err)
	})
}
