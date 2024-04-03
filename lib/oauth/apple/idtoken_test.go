package apple_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"github.com/keratin/authn-server/lib/oauth/apple"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type mockKeyStore struct {
	keys map[string]*rsa.PublicKey
}

func (ks *mockKeyStore) Get(keyID string) (*rsa.PublicKey, error) {
	if key, ok := ks.keys[keyID]; ok {
		return key, nil
	}
	return nil, &apple.KeyNotFoundError{KeyID: keyID}
}

func TestGetUserDetailsFromToken(t *testing.T) {
	testAppleSigningKey := func() (jose.SigningKey, string, *rsa.PrivateKey) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		keyID := uuid.NewString()

		signingKey := jose.SigningKey{Key: jose.JSONWebKey{Key: rsaKey, KeyID: keyID, Algorithm: string(jose.RS256), Use: "sig"}, Algorithm: jose.RS256}

		return signingKey, keyID, rsaKey
	}

	t.Run("valid", func(t *testing.T) {
		signingKey, keyID, privKey := testAppleSigningKey()
		clientID := uuid.NewString()
		sut := apple.NewTokenReader(clientID, apple.WithKeyStore(&mockKeyStore{keys: map[string]*rsa.PublicKey{keyID: &privKey.PublicKey}}))

		claims := apple.Claims{
			Claims: jwt.Claims{
				Issuer:   apple.BaseURL,
				Audience: jwt.Audience{clientID},
				Expiry:   jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
				IssuedAt: jwt.NewNumericDate(time.Now().Add(-5 * time.Second)),
				Subject:  uuid.NewString(),
			},
			Email: "claimed@example.com",
		}

		signer, err := jose.NewSigner(signingKey, &jose.SignerOptions{})
		require.NoError(t, err)

		idToken, err := jwt.Signed(signer).
			Claims(claims).
			CompactSerialize()
		require.NoError(t, err)

		tok := &oauth2.Token{}
		tok = tok.WithExtra(map[string]interface{}{"id_token": idToken})

		id, email, err := sut.GetUserDetailsFromToken(tok)

		assert.NoError(t, err)
		assert.Equal(t, claims.Subject, id)
		assert.Equal(t, claims.Email, email)
	})

	t.Run("invalid", func(t *testing.T) {
		sut := apple.NewTokenReader("", apple.WithKeyStore(&mockKeyStore{keys: map[string]*rsa.PublicKey{}}))

		t.Run("missing", func(t *testing.T) {
			tok := &oauth2.Token{}
			_, _, err := sut.GetUserDetailsFromToken(tok)
			assert.EqualError(t, err, "missing id_token")
		})

		t.Run("not a string", func(t *testing.T) {
			tok := &oauth2.Token{}
			tok = tok.WithExtra(map[string]interface{}{"id_token": 0})
			_, _, err := sut.GetUserDetailsFromToken(tok)
			assert.EqualError(t, err, "id_token is not a string")
		})

		t.Run("failed to parse", func(t *testing.T) {
			tok := &oauth2.Token{}
			tok = tok.WithExtra(map[string]interface{}{"id_token": "not a jwt"})
			_, _, err := sut.GetUserDetailsFromToken(tok)
			assert.EqualError(t, err, "go-jose/go-jose: compact JWS format must have three parts")
		})

		t.Run("invalid key header", func(t *testing.T) {
			signingKey, _, _ := testAppleSigningKey()

			claims := map[string]interface{}{
				"iss": apple.BaseURL,
			}

			signer, err := jose.NewSigner(signingKey, &jose.SignerOptions{
				ExtraHeaders: map[jose.HeaderKey]interface{}{
					"alg": jose.ES256, // wrong algorithm won't be found
				},
			})
			require.NoError(t, err)

			idToken, err := jwt.Signed(signer).
				Claims(claims).
				CompactSerialize()
			require.NoError(t, err)

			tok := &oauth2.Token{}
			tok = tok.WithExtra(map[string]interface{}{"id_token": idToken})

			_, _, err = sut.GetUserDetailsFromToken(tok)
			assert.EqualError(t, err, "no RS256 key header found")
		})

		t.Run("key not found", func(t *testing.T) {
			// key is NOT added to mock key store
			signingKey, keyID, _ := testAppleSigningKey()

			claims := map[string]interface{}{
				"iss": apple.BaseURL,
			}

			signer, err := jose.NewSigner(signingKey, &jose.SignerOptions{})
			require.NoError(t, err)

			idToken, err := jwt.Signed(signer).
				Claims(claims).
				CompactSerialize()
			require.NoError(t, err)

			tok := &oauth2.Token{}
			tok = tok.WithExtra(map[string]interface{}{"id_token": idToken})

			_, _, err = sut.GetUserDetailsFromToken(tok)

			var expectedErr *apple.KeyNotFoundError
			assert.True(t, errors.As(err, &expectedErr))
			assert.Equal(t, keyID, expectedErr.KeyID)
		})
	})
}
