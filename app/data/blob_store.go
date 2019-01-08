package data

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	dataRedis "github.com/keratin/authn-server/app/data/redis"
	"github.com/keratin/authn-server/app/data/sqlite3"
	"github.com/keratin/authn-server/ops"
)

type BlobStore interface {
	// Read fetches a blob from the store.
	Read(name string) ([]byte, error)

	// WriteNX will write the blob into the store only if the name does not exist.
	WriteNX(name string, blob []byte) (bool, error)
}

func NewBlobStore(interval time.Duration, redis *redis.Client, db *sqlx.DB, reporter ops.ErrorReporter) (BlobStore, error) {
	// the lifetime of a key should be slightly more than two intervals
	ttl := interval*2 + 10*time.Second

	// the write lock should be greater than the peak time necessary to generate and encrypt a key,
	// plus send it back over the wire to redis. after this time has elapsed, any other authn server
	// may get the lock and generate a competing key.
	lockTime := time.Duration(1500) * time.Millisecond

	if redis != nil {
		return &dataRedis.BlobStore{
			TTL:      ttl,
			LockTime: lockTime,
			Client:   redis,
		}, nil
	}

	switch db.DriverName() {
	case "sqlite3":
		store := &sqlite3.BlobStore{
			TTL:      ttl,
			LockTime: lockTime,
			DB:       db,
		}
		store.Clean(reporter)
		return store, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %v", db.DriverName())
	}
}
