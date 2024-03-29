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

/*
func TestExtractUserFromClaims(t *testing.T) {
	clientID := uuid.NewString()

	t.Run("valid", func(t *testing.T) {
		appleUserID := uuid.NewString()
		email := uuid.NewString()

		claims := map[string]interface{}{
			"iss":   BaseURL,
			"aud":   clientID,
			"exp":   time.Now().Unix() + 5,
			"sub":   appleUserID,
			"email": email,
		}

		id, e, err := extractUserFromClaims(claims, clientID)
		assert.NoError(t, err)
		assert.Equal(t, appleUserID, id)
		assert.Equal(t, email, e)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("issuer", func(t *testing.T) {
			claims := map[string]interface{}{
				"iss": "https://example.com",
			}
			_, _, err := extractUserFromClaims(claims, clientID)
			assert.EqualError(t, err, "invalid or missing issuer")
		})

		t.Run("audience", func(t *testing.T) {
			claims := map[string]interface{}{
				"iss": BaseURL,
				"aud": "not the client id",
			}
			_, _, err := extractUserFromClaims(claims, clientID)
			assert.EqualError(t, err, "invalid or missing audience")
		})

		t.Run("expiration", func(t *testing.T) {
			t.Run("non-numeric", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss": BaseURL,
					"aud": clientID,
					"exp": "not a number",
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "invalid exp")
			})

			t.Run("float64", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss": BaseURL,
					"aud": clientID,
					"exp": 0.0,
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "token expired")
			})

			t.Run("int", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss": BaseURL,
					"aud": clientID,
					"exp": 0,
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "token expired")
			})
		})

		t.Run("sub", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss": BaseURL,
					"aud": clientID,
					"exp": time.Now().Unix() + 5,
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "missing claim 'sub'")
			})

			t.Run("not a string", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss": BaseURL,
					"aud": clientID,
					"exp": time.Now().Unix() + 5,
					"sub": 0,
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "claim 'sub' is not a string")
			})
		})

		t.Run("email", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss": BaseURL,
					"aud": clientID,
					"exp": time.Now().Unix() + 5,
					"sub": "user id",
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "missing claim 'email'")
			})

			t.Run("not a string", func(t *testing.T) {
				claims := map[string]interface{}{
					"iss":   BaseURL,
					"aud":   clientID,
					"exp":   time.Now().Unix() + 5,
					"sub":   "user id",
					"email": 0,
				}
				_, _, err := extractUserFromClaims(claims, clientID)
				assert.EqualError(t, err, "claim 'email' is not a string")
			})
		})
	})
}

func TestValidateExp(t *testing.T) {
	for _, tc := range []struct {
		name    string
		expired interface{}
		ok      interface{}
	}{
		{"float64", 0.0, float64(time.Now().Unix() + 5)},
		{"int", 0, time.Now().Unix() + 5},
		{"int32", int32(0), int32(time.Now().Unix() + 5)},
		{"int64", int64(0), int64(time.Now().Unix() + 5)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("expired", func(t *testing.T) {
				err := validateExp(tc.expired)
				assert.EqualError(t, err, "token expired")
			})

			t.Run("ok", func(t *testing.T) {
				err := validateExp(tc.ok)
				assert.NoError(t, err)
			})
		})
	}
}
*/
