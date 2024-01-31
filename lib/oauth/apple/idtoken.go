package apple

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"golang.org/x/oauth2"
)

type TokenReader struct {
	keyStore rsaKeyStore
	clientID string
}

func NewTokenReader(clientID string) *TokenReader {
	return &TokenReader{
		clientID: clientID,
		keyStore: newSigningKeyStore(&http.Client{
			Timeout: 10 * time.Second,
		}),
	}
}

func (tr *TokenReader) GetUserDetailsFromToken(t *oauth2.Token) (string, string, error) {
	claims, err := tr.getAppleIDTokenClaims(t)
	if err != nil {
		return "", "", fmt.Errorf("failed to get apple ID token claims: %w", err)
	}

	return extractUserFromClaims(claims, tr.clientID)
}

func (tr *TokenReader) getAppleIDTokenClaims(t *oauth2.Token) (map[string]interface{}, error) {
	idTokenVal := t.Extra("id_token")
	if idTokenVal == nil {
		return nil, fmt.Errorf("missing id_token")
	}

	idToken, ok := idTokenVal.(string)

	if !ok {
		return nil, fmt.Errorf("id_token is not a string")
	}

	parsedIDToken, err := jwt.ParseSigned(idToken)

	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("no RS256 key header found")
	}

	appleRSA, err := tr.keyStore.get(hdr.KeyID)

	if err != nil {
		return nil, fmt.Errorf("failed to get apple RSA key: %w", err)
	}

	claims := make(map[string]interface{})
	err = parsedIDToken.Claims(appleRSA, &claims)

	if err != nil {
		return nil, fmt.Errorf("failed to verify claims: %w", err)
	}

	return claims, nil
}

func extractUserFromClaims(claims map[string]interface{}, clientID string) (string, string, error) {
	// We could validate iat here if we had a good minimum value to use.
	// A nonce claim is also available but would need to be sent on code exchange.
	if iss, ok := claims["iss"]; !ok || !strings.Contains(iss.(string), BaseURL) {
		return "", "", fmt.Errorf("invalid or missing issuer")
	}

	if aud, ok := claims["aud"]; !ok || aud.(string) != clientID {
		return "", "", fmt.Errorf("invalid or missing audience")
	}

	if exp, ok := claims["exp"]; !ok {
		return "", "", fmt.Errorf("missing exp")
	} else {
		expErr := validateExp(exp)
		if expErr != nil {
			return "", "", expErr
		}
	}

	id, ok := claims["sub"]

	if !ok {
		return "", "", fmt.Errorf("missing claim 'sub'")
	}

	idString, ok := id.(string)
	if !ok {
		return "", "", fmt.Errorf("claim 'sub' is not a string")
	}

	email, ok := claims["email"]

	if !ok {
		return "", "", fmt.Errorf("missing claim 'email'")
	}

	emailString, ok := email.(string)
	if !ok {
		return "", "", fmt.Errorf("claim 'email' is not a string")
	}

	return idString, emailString, nil
}

func validateExp(exp interface{}) error {
	switch v := exp.(type) {
	case float64:
		return validateExpInt64(int64(v))
	case int:
		return validateExpInt64(int64(v))
	case int32:
		return validateExpInt64(int64(v))
	case int64:
		return validateExpInt64(v)
	default:
		return fmt.Errorf("invalid exp")
	}
}

func validateExpInt64(exp int64) error {
	if exp < time.Now().Unix() {
		return fmt.Errorf("token expired")
	}

	return nil
}
