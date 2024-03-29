package apple

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"golang.org/x/oauth2"
)

type TokenReader struct {
	keyStore rsaKeyStore
	clientID string
}

func NewTokenReader(clientID string, opts ...func(*TokenReader)) *TokenReader {
	tr := &TokenReader{
		clientID: clientID,
	}

	for _, opt := range opts {
		opt(tr)
	}

	if tr.keyStore == nil {
		tr.keyStore = newSigningKeyStore(&http.Client{
			Timeout: 10 * time.Second,
		})
	}

	return tr
}

func WithKeyStore(ks rsaKeyStore) func(*TokenReader) {
	return func(tr *TokenReader) {
		tr.keyStore = ks
	}
}

func (tr *TokenReader) GetUserDetailsFromToken(t *oauth2.Token) (string, string, error) {
	idTokenVal := t.Extra("id_token")
	if idTokenVal == nil {
		return "", "", fmt.Errorf("missing id_token")
	}

	idToken, ok := idTokenVal.(string)
	if !ok {
		return "", "", fmt.Errorf("id_token is not a string")
	}

	parsedIDToken, err := jwt.ParseSigned(idToken)
	if err != nil {
		return "", "", err
	}

	var hdr *jose.Header
	for i := range parsedIDToken.Headers {
		th := parsedIDToken.Headers[i]
		if th.Algorithm == "RS256" {
			hdr = &th
			break
		}
	}
	if hdr == nil {
		return "", "", fmt.Errorf("no RS256 key header found")
	}

	appleRSA, err := tr.keyStore.Get(hdr.KeyID)
	if err != nil {
		return "", "", fmt.Errorf("failed to Get apple RSA key: %w", err)
	}

	claims := Claims{}
	err = parsedIDToken.Claims(appleRSA, &claims)
	if err != nil {
		return "", "", fmt.Errorf("failed to verify claims: %w", err)
	}

	// TODO: figure out a way to cleanly pass the nonce in authorize request and make available for validation here
	err = claims.Validate(tr.clientID)

	if err != nil {
		return "", "", fmt.Errorf("failed to validate claims: %w", err)
	}
	return claims.Subject, claims.Email, nil
}
