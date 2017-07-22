package mock

import "crypto/rsa"

type keyStore struct {
	key *rsa.PrivateKey
}

func NewKeyStore(key *rsa.PrivateKey) *keyStore {
	return &keyStore{key}
}

func (ks *keyStore) Key() *rsa.PrivateKey {
	return ks.key
}

func (ks *keyStore) Keys() []*rsa.PrivateKey {
	return []*rsa.PrivateKey{ks.key}
}
