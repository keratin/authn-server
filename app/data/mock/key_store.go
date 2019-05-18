package mock

import (
	"github.com/keratin/authn-server/app/data/private"
)

type keyStore struct {
	key *private.Key
}

func NewKeyStore(key *private.Key) *keyStore {
	return &keyStore{key}
}

func (ks *keyStore) Key() *private.Key {
	return ks.key
}

func (ks *keyStore) Keys() []*private.Key {
	return []*private.Key{ks.key}
}
