package services_test

import (
	"testing"

	"github.com/keratin/authn-server/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/keratin/authn-server/data/mock"
)

func TestLastLoginUpdater(t *testing.T) {
	t.Run("last login", func(t *testing.T) {
		accountStore := mock.NewAccountStore()
		existing, err := accountStore.Create("existing", []byte("secret"))
		require.NoError(t, err)

		err = services.LastLoginUpdater(accountStore, existing.ID)
		require.NoError(t, err)

		found, err := accountStore.Find(existing.ID)
		require.NoError(t, err)
		assert.NotEqual(t, nil, found.LastLoginAt)
	})
}
