package compat

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	jose "github.com/square/go-jose"
)

func KeyID(key crypto.PublicKey) string {
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		panic(fmt.Errorf("Not a RSA key"))
	}

	jwk := jose.JSONWebKey{Key: rsaKey}
	kid, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(kid)
}
