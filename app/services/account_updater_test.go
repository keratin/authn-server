package services_test

import (
	"testing"

	"github.com/keratin/authn-server/app/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
)

func TestAccountUpdater(t *testing.T) {
	accountStore := mock.NewAccountStore()
	existing, err := accountStore.Create("existing", []byte("secret"))
	require.NoError(t, err)

	t.Run("email usernames", func(t *testing.T) {
		cfg := &app.Config{
			UsernameIsEmail: true,
		}

		t.Run("success", func(t *testing.T) {
			err := services.AccountUpdater(accountStore, cfg, existing.ID, "new@email.tech")
			require.NoError(t, err)

			found, err := accountStore.Find(existing.ID)
			require.NoError(t, err)
			assert.Equal(t, "new@email.tech", found.Username)
		})

		t.Run("invalid", func(t *testing.T) {
			err := services.AccountUpdater(accountStore, cfg, existing.ID, "invalid")
			assert.Equal(t, services.FieldErrors{{"username", services.ErrFormatInvalid}}, err)
		})
	})

	t.Run("username taken", func(t *testing.T) {
		cfg := &app.Config{
			UsernameIsEmail:   false,
			UsernameMinLength: 3,
		}

		other, err := accountStore.Create("other", []byte("secret"))
		require.NoError(t, err)

		err = services.AccountUpdater(accountStore, cfg, existing.ID, other.Username)
		assert.Equal(t, services.FieldErrors{{"username", services.ErrTaken}}, err)
	})

	t.Run("string usernames", func(t *testing.T) {
		cfg := &app.Config{
			UsernameIsEmail:   false,
			UsernameMinLength: 5,
		}

		t.Run("success", func(t *testing.T) {
			err := services.AccountUpdater(accountStore, cfg, existing.ID, "newname")
			require.NoError(t, err)

			found, err := accountStore.Find(existing.ID)
			require.NoError(t, err)
			assert.Equal(t, "newname", found.Username)
		})

		t.Run("invalid", func(t *testing.T) {
			err := services.AccountUpdater(accountStore, cfg, existing.ID, "nope")
			assert.Equal(t, services.FieldErrors{{"username", services.ErrFormatInvalid}}, err)
		})
	})
}
