package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/ops"
	"github.com/pkg/errors"
)

// NewKeyStore creates a key store that uses Redis to persist an auto-generated key and rotate it
// regularly. The key is encrypted using SECRET_KEY_BASE, which is already the ultimate SPOF for
// AuthN security. It's expected that very few people will be in position to improve on the security
// tradeoffs of this provider.
func NewKeyStore(client *redis.Client, reporter ops.ErrorReporter, interval time.Duration, race time.Duration, encryptionKey []byte) (*data.RotatingKeyStore, error) {
	ks := data.NewRotatingKeyStore()

	m := &maintainer{
		store: &BlobStore{
			// the lifetime of a key should be slightly more than two intervals
			TTL: interval*2 + 10*time.Second,
			// this should be greater than the peak time necessary to generate and encrypt a
			// key, plus send it back over the wire to redis.
			LockTime: race,
			Client:   client,
		},
		// the rotation interval should be slightly longer than access token expiry.
		// this means that when a key goes inactive for some interval, we can know
		// that it is useless and discardable by the third interval.
		interval:      interval,
		keyStrength:   2048,
		encryptionKey: encryptionKey,
	}
	err := m.maintain(ks, reporter)
	if err != nil {
		return nil, errors.Wrap(err, "maintain")
	}

	return ks, nil
}
