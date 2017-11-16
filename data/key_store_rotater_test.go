package data_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyStoreRotater(t *testing.T) {
	reporter := &ops.LogReporter{}
	secret := []byte("32bigbytesofsuperultimatesecrecy")
	interval := time.Hour

	t.Run("empty remote storage", func(t *testing.T) {
		blobStore := data.NewEncryptedBlobStore(mock.NewBlobStore(interval*2+time.Second, time.Second), secret)
		store := data.NewRotatingKeyStore()
		rotater := data.NewKeyStoreRotater(blobStore, interval)
		err := rotater.Maintain(store, reporter)
		require.NoError(t, err)

		assert.NotEmpty(t, store.Keys())
		assert.Len(t, store.Keys(), 1)
		assert.Equal(t, store.Key(), store.Keys()[0])
	})

	t.Run("multiple servers", func(t *testing.T) {
		blobStore := data.NewEncryptedBlobStore(mock.NewBlobStore(interval*2+time.Second, time.Second), secret)

		store1 := data.NewRotatingKeyStore()
		err := data.NewKeyStoreRotater(blobStore, interval).Maintain(store1, reporter)
		require.NoError(t, err)
		key1 := store1.Key()
		assert.NotEmpty(t, key1)

		store2 := data.NewRotatingKeyStore()
		err = data.NewKeyStoreRotater(blobStore, interval).Maintain(store2, reporter)
		require.NoError(t, err)
		assert.Len(t, store2.Keys(), 1)
		assert.Equal(t, key1, store2.Key())
		assert.Equal(t, key1, store2.Keys()[0])
	})

	t.Run("rotation", func(t *testing.T) {
		blobStore := data.NewEncryptedBlobStore(mock.NewBlobStore(interval*2+time.Second, time.Second), secret)
		store := data.NewRotatingKeyStore()
		rotater := data.NewKeyStoreRotater(blobStore, interval)
		err := rotater.Maintain(store, reporter)
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
