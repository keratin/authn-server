package data

import (
	"sync"

	"github.com/keratin/authn-server/app/data/private"
)

// RotatingKeyStore is a KeyStore that may be rotated by a maintainer.
type RotatingKeyStore struct {
	keys   []*private.Key
	rwLock *sync.RWMutex
}

// NewRotatingKeyStore builds a RotatingKeyStore
func NewRotatingKeyStore() *RotatingKeyStore {
	return &RotatingKeyStore{
		keys:   []*private.Key{},
		rwLock: &sync.RWMutex{},
	}
}

// Key returns the current key. It relies on the internal keys slice being sorted with the newest
// key last.
func (ks *RotatingKeyStore) Key() *private.Key {
	ks.rwLock.RLock()
	defer ks.rwLock.RUnlock()

	if len(ks.keys) > 0 {
		return ks.keys[len(ks.keys)-1]
	} else {
		return nil
	}
}

// Keys will return the previous and current keys, in that order.
func (ks *RotatingKeyStore) Keys() []*private.Key {
	ks.rwLock.RLock()
	defer ks.rwLock.RUnlock()

	return ks.keys
}

// Rotate is responsible for adding a new key to the list. It maintains key order from oldest to
// newest, and ensures a maximum of two entries.
func (ks *RotatingKeyStore) Rotate(k *private.Key) {
	var keys []*private.Key
	if len(ks.keys) > 0 {
		keys = append(keys, ks.keys[len(ks.keys)-1])
	}
	keys = append(keys, k)

	ks.rwLock.Lock()
	defer ks.rwLock.Unlock()
	ks.keys = keys
}
