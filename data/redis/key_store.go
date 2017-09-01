package redis

import (
	"crypto/rsa"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var placeholder = "generating"

type keyStore struct {
	keys   []*rsa.PrivateKey
	rwLock *sync.RWMutex
}

// NewKeyStore creates a key store that uses Redis to persist an auto-generated key and rotate it
// regularly. The key is encrypted using SECRET_KEY_BASE, which is already the ultimate SPOF for
// AuthN security. It's expected that very few people will be in position to improve on the security
// tradeoffs of this provider.
func NewKeyStore(client *redis.Client, interval time.Duration, race time.Duration, encryptionKey []byte) (*keyStore, error) {
	ks := &keyStore{
		keys:   []*rsa.PrivateKey{},
		rwLock: &sync.RWMutex{},
	}

	m := &maintainer{
		interval:      interval,
		race:          race,
		keyStrength:   2048,
		client:        client,
		encryptionKey: encryptionKey,
	}
	err := m.maintain(ks)
	if err != nil {
		return nil, errors.Wrap(err, "maintain")
	}

	return ks, nil
}

// Key returns the current key. It relies on the internal keys slice being sorted with the newest
// key last.
func (ks *keyStore) Key() *rsa.PrivateKey {
	ks.rwLock.RLock()
	defer ks.rwLock.RUnlock()

	return ks.keys[len(ks.keys)-1]
}

// Keys will return the previous and current keys, in that order.
func (ks *keyStore) Keys() []*rsa.PrivateKey {
	ks.rwLock.RLock()
	defer ks.rwLock.RUnlock()

	return ks.keys
}

// Rotate is responsible for adding a new key to the list. It maintains key order from oldest to
// newest, and ensures a maximum of two entries.
func (ks *keyStore) Rotate(k *rsa.PrivateKey) {
	keys := []*rsa.PrivateKey{}
	if len(ks.keys) > 0 {
		keys = append(keys, ks.keys[len(ks.keys)-1])
	}
	keys = append(keys, k)

	ks.rwLock.Lock()
	defer ks.rwLock.Unlock()
	ks.keys = keys
}
