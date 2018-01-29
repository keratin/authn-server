package compat

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
)

// KeyID uses square/go-jose to extract the JWK thumbprint for a RSA public key.
func KeyID(key crypto.PublicKey) (string, error) {
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("Not a RSA key")
	}

	jwk := jose.JSONWebKey{Key: rsaKey}
	kid, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", errors.Wrap(err, "jwk.Thumbprint")
	}

	return base64.RawURLEncoding.EncodeToString(kid), nil
}
