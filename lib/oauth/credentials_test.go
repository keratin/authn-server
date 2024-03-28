package oauth

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewCredentials(t *testing.T) {
	t.Run("invalid credentials", func(t *testing.T) {
		credentials, err := NewCredentials("id")
		assert.NotNil(t, err)
		assert.Equal(t, "credentials must be in the format `id:secret:additional=data...(optional)`", err.Error())
		assert.Nil(t, credentials)
	})

	t.Run("valid credentials", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s", id, secret))
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
	})

	t.Run("valid credentials with additional", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s:e1=extra:e2=data", id, secret))
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
		assert.Equal(t, map[string]string{"e1": "extra", "e2": "data"}, credentials.Additional)
	})

	t.Run("valid credentials with empty additional value", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s:e1:e2=data", id, secret))
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
		assert.Equal(t, map[string]string{"e1": "", "e2": "data"}, credentials.Additional)
	})

	t.Run("valid credentials with empty additional pair", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s::e1=extra::e2=data", id, secret))
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
		assert.Equal(t, map[string]string{"e1": "extra", "e2": "data"}, credentials.Additional)
	})

	t.Run("valid credentials with signing key", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()
		signingKey := []byte("key-override-a-reno")

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s:%s", id, secret, hex.EncodeToString(signingKey)))
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
	})
}
