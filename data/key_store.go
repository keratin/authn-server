package data

import (
	"crypto/rsa"
)

type KeyStore interface {
	// Returns the current key
	Key() *rsa.PrivateKey
	// Returns recent keys (including current key)
	Keys() []*rsa.PrivateKey
}
