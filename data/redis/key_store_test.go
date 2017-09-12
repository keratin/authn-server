package redis_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/keratin/authn-server/data/redis"
	"github.com/keratin/authn-server/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyStore(t *testing.T) {
	client, err := redis.TestDB()
	reporter := &ops.LogReporter{}
	require.NoError(t, err)
	secret := []byte("32bigbytesofsuperultimatesecrecy")

	t.Run("empty remote storage", func(t *testing.T) {
		client.FlushDB()
		store, err := redis.NewKeyStore(client, reporter, time.Hour, time.Second, secret)
		require.NoError(t, err)

		assert.NotEmpty(t, store.Keys())
		assert.Len(t, store.Keys(), 1)
		assert.Equal(t, store.Key(), store.Keys()[0])
	})

	t.Run("multiple servers", func(t *testing.T) {
		client.FlushDB()
		store1, err := redis.NewKeyStore(client, reporter, time.Hour, time.Second, secret)
		require.NoError(t, err)
		key1 := store1.Key()
		assert.NotEmpty(t, key1)

		store2, err := redis.NewKeyStore(client, reporter, time.Hour, time.Second, secret)
		require.NoError(t, err)
		assert.Equal(t, key1, store2.Key())
		assert.Len(t, store2.Keys(), 1)
		assert.Equal(t, key1, store2.Keys()[0])
	})

	t.Run("rotation", func(t *testing.T) {
		client.FlushDB()
		store, err := redis.NewKeyStore(client, reporter, time.Hour, time.Second, secret)
		require.NoError(t, err)

		firstKey := store.Keys()[0]

		secondKey, err := rsa.GenerateKey(rand.Reader, 256)
		require.NoError(t, err)
		store.Rotate(secondKey)
		assert.Equal(t, []*rsa.PrivateKey{firstKey, secondKey}, store.Keys())

		thirdKey, err := rsa.GenerateKey(rand.Reader, 256)
		require.NoError(t, err)
		store.Rotate(thirdKey)
		assert.Equal(t, []*rsa.PrivateKey{secondKey, thirdKey}, store.Keys())
	})
}
