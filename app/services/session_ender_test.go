package services_test

import (
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSessionEnder(t *testing.T) {
	accountID := 0
	refreshStore := mock.NewRefreshTokenStore()

	t.Run("revokes token", func(t *testing.T) {
		token, err := refreshStore.Create(accountID)
		require.NoError(t, err)

		err = services.SessionEnder(refreshStore, &token)
		assert.NoError(t, err)

		foundID, err := refreshStore.Find(token)
		assert.Empty(t, foundID)
		assert.NoError(t, err)
	})

	t.Run("ignores missing token", func(t *testing.T) {
		err := services.SessionEnder(refreshStore, nil)
		assert.NoError(t, err)
	})
}
