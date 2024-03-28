package apple

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

// ParseAppleKey takes a PEM encoded ES256 key and returns a jose.SigningKey.
func ParsePrivateKey(keyBytes []byte, keyID string) (*jose.SigningKey, error) {
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode PEM data")
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC key: %w", err)
	}

	return &jose.SigningKey{Key: jose.JSONWebKey{
		Key:       key,
		KeyID:     keyID,
		Algorithm: "ES256",
		Use:       "sig",
	}, Algorithm: jose.ES256}, nil
}

// GenerateSecret creates a signed JWT as specified at
// https://developer.apple.com/documentation/accountorganizationaldatasharing/creating-a-client-secret
func GenerateSecret(key jose.SigningKey, keyID, clientID, teamID string, expiresInSeconds int64) (string, error) {
	if key.Algorithm != jose.ES256 {
		return "", fmt.Errorf("expected ES256 signing key got %s", key.Algorithm)
	}

	signer, err := jose.NewSigner(key, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"kid": keyID,
			"alg": key.Algorithm,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	return jwt.Signed(signer).Claims(map[string]interface{}{
		"iss": teamID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(expiresInSeconds) * time.Second).Unix(),
		"aud": BaseURL,
		"sub": clientID,
	}).CompactSerialize() // TODO: compact or full serialization?
}
