package redis_test

import (
	"testing"
	"time"

	goredis "github.com/go-redis/redis"
	"github.com/keratin/authn-server/data/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyStore(t *testing.T) {
	client := goredis.NewClient(&goredis.Options{
		Addr: "127.0.0.1:6379",
		DB:   12,
	})
	secret := []byte("32bigbytesofsuperultimatesecrecy")

	t.Run("empty remote storage", func(t *testing.T) {
		client.FlushDB()
		store := redis.NewKeyStore(client, time.Hour, time.Second, secret)

		keys, err := store.Keys()
		require.NoError(t, err)
		assert.Empty(t, keys)

		key, err := store.Key()
		require.NoError(t, err)
		assert.NotEmpty(t, key)

		keys, err = store.Keys()
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, key, keys[0])
	})

	t.Run("recovering an existing key", func(t *testing.T) {
		client.FlushDB()
		store1 := redis.NewKeyStore(client, time.Hour, time.Second, secret)
		key1, err := store1.Key() // trigger findOrCreate
		require.NoError(t, err)

		store2 := redis.NewKeyStore(client, time.Hour, time.Second, secret)
		key2, err := store2.Key()
		require.NoError(t, err)
		assert.Equal(t, key1, key2)

		store3 := redis.NewKeyStore(client, time.Hour, time.Second, secret)
		keys3, err := store3.Keys()
		require.NoError(t, err)
		assert.Len(t, keys3, 1)
		assert.Equal(t, key1, keys3[0])
	})
}
