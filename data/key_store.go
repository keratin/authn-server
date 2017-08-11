package data

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type KeyStore interface {
	// Returns the current key
	Key() *rsa.PrivateKey
	// Returns recent keys (including current key)
	Keys() []*rsa.PrivateKey
}

func RSAPublicKeyID(key crypto.PublicKey) string {
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		panic(fmt.Errorf("Not a RSA key"))
	}

	json, err := json.Marshal(map[string]interface{}{
		"Kty": "RSA",
		"N":   rsaKey.N,
		"E":   rsaKey.E,
	})
	if err != nil {
		panic(err)
	}

	fingerprint := sha256.Sum256(json)
	return base64.RawURLEncoding.EncodeToString(fingerprint[:])
}
