package data

import (
	"time"

	"github.com/go-redis/redis"
	dataRedis "github.com/keratin/authn-server/data/redis"
)

type BlobStore interface {
	// Read fetches a blob from the store.
	Read(name string) ([]byte, error)

	// WLock acquires a global mutex that will either timeout or be
	// released by a successful Write
	WLock(name string) (bool, error)

	// Write puts a blob into the store.
	Write(name string, blob []byte) error
}

func NewBlobStore(interval time.Duration, redis *redis.Client) BlobStore {
	return &dataRedis.BlobStore{
		// the lifetime of a key should be slightly more than two intervals
		TTL: interval*2 + 10*time.Second,
		// the write lock should be greater than the peak time necessary to generate and
		// encrypt a key, plus send it back over the wire to redis.
		LockTime: time.Duration(500) * time.Millisecond,
		Client:   redis,
	}
}
