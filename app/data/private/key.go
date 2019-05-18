package private

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

type Key struct {
	JWK jose.JSONWebKey
	*rsa.PrivateKey
}

// Wrap the provided RSA private key as our internal key with canonical ID
func NewKey(key *rsa.PrivateKey) (*Key, error) {
	id, err := keyID(&key.PublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "private.keyID")
	}
	return &Key{
		PrivateKey: key,
		JWK: jose.JSONWebKey{
			Key:       key.Public(),
			Use:       "sig",
			Algorithm: "RS256",
			KeyID:     id,
		},
	}, nil
}

// Generate a bits wide RSA private key
func GenerateKey(bits int) (*Key, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return NewKey(key)
}

// KeyID uses square/go-jose to extract the JWK thumbprint for a RSA public key.
func keyID(key *rsa.PublicKey) (string, error) {
	jwk := jose.JSONWebKey{Key: key}
	kid, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", errors.Wrap(err, "jwk.Thumbprint")
	}

	return base64.RawURLEncoding.EncodeToString(kid), nil
}
