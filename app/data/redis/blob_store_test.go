package redis_test

import (
	"testing"
	"time"

	"github.com/keratin/authn-server/app/data/redis"
	"github.com/keratin/authn-server/app/data/testers"
	"github.com/stretchr/testify/require"
)

func TestBlobStore(t *testing.T) {
	client, err := redis.TestDB()
	require.NoError(t, err)
	store := &redis.BlobStore{
		Client:   client,
		TTL:      time.Second,
		LockTime: time.Second,
	}
	for _, tester := range testers.BlobStoreTesters {
		tester(t, store)
		client.FlushDB()
	}
}
