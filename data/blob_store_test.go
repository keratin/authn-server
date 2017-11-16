package data_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/data/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlobStores(t *testing.T) {
	testers := []func(*testing.T, data.BlobStore){
		testReadWrite,
		testWriteLock,
	}

	t.Run("mock", func(t *testing.T) {
		for _, tester := range testers {
			store := mock.NewBlobStore(time.Second, time.Second)
			tester(t, store)
		}
	})

	t.Run("Redis", func(t *testing.T) {
		client, err := redis.TestDB()
		require.NoError(t, err)
		store := &redis.BlobStore{
			Client:   client,
			TTL:      time.Second,
			LockTime: time.Second,
		}
		for _, tester := range testers {
			tester(t, store)
			client.FlushDb()
		}
	})
}

func testReadWrite(t *testing.T, bs data.BlobStore) {
	blob, err := bs.Read("unknown")
	assert.NoError(t, err)
	assert.Empty(t, blob)

	err = bs.Write("blob", []byte("val"))
	assert.NoError(t, err)

	blob, err = bs.Read("blob")
	assert.NoError(t, err)
	assert.Equal(t, []byte("val"), blob)
}

func testWriteLock(t *testing.T, bs data.BlobStore) {
	ok, err := bs.WLock("lockedKey")
	assert.NoError(t, err)
	assert.True(t, ok)

	// can't re-acquire, even with the same connection
	ok, err = bs.WLock("lockedKey")
	assert.NoError(t, err)
	assert.False(t, ok)

	// write lock does not create findable blob
	blob, err := bs.Read("lockedKey")
	assert.NoError(t, err)
	assert.Empty(t, blob)

	// can't lock an existing key
	err = bs.Write("existing", []byte("val"))
	assert.NoError(t, err)
	ok, err = bs.WLock("existing")
	assert.NoError(t, err)
	assert.False(t, ok)

}
