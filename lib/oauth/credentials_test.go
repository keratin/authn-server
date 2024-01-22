package oauth

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewCredentials(t *testing.T) {
	defaultKey := []byte("key-a-reno")

	t.Run("invalid credentials", func(t *testing.T) {
		credentials, err := NewCredentials("id", defaultKey)
		assert.NotNil(t, err)
		assert.Equal(t, "Credentials must be in the format `id:string:signing_key(optional)`", err.Error())
		assert.Nil(t, credentials)
	})

	t.Run("valid credentials", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s", id, secret), defaultKey)
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
		assert.Equal(t, defaultKey, credentials.SigningKey)
	})

	t.Run("valid credentials with signing key", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()
		signingKey := []byte("key-override-a-reno")

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s:%s", id, secret, hex.EncodeToString(signingKey)), defaultKey)
		assert.Nil(t, err)
		assert.NotNil(t, credentials)
		assert.Equal(t, id, credentials.ID)
		assert.Equal(t, secret, credentials.Secret)
		assert.Equal(t, signingKey, credentials.SigningKey)
	})

	t.Run("invalid signing key", func(t *testing.T) {
		id := uuid.NewString()
		secret := uuid.NewString()
		badKey := fmt.Sprintf("g%s", uuid.NewString()) // g is not a valid hex character

		credentials, err := NewCredentials(fmt.Sprintf("%s:%s:%s", id, secret, badKey), defaultKey)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to decode signing key: encoding/hex: invalid byte: U+0067 'g'", err.Error())
		assert.Nil(t, credentials)
	})
}
