package apple_test

import (
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/keratin/authn-server/lib/oauth/apple"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestClaimsValidate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("empty clientID", func(t *testing.T) {
			claims := apple.Claims{}
			err := claims.Validate("")
			assert.EqualError(t, err, "cannot validate with empty clientID")
		})

		t.Run("missing email", func(t *testing.T) {
			claims := apple.Claims{}
			err := claims.Validate("audience")
			assert.EqualError(t, err, "missing claim 'email'")
		})

		t.Run("missing Expiry", func(t *testing.T) {
			claims := apple.Claims{
				Email:  "email",
				Claims: jwt.Claims{},
			}
			err := claims.Validate("audience")
			assert.EqualError(t, err, "missing claim 'exp'")
		})

		t.Run("missing IssuedAt", func(t *testing.T) {
			claims := apple.Claims{
				Email: "email",
				Claims: jwt.Claims{
					Expiry: jwt.NewNumericDate(time.Now()),
				},
			}
			err := claims.Validate("audience")
			assert.EqualError(t, err, "missing claim 'iat'")
		})
	})

	// ensure we don't break underlying jwt.Claims validation
	t.Run("jwt.Claims.Validate", func(t *testing.T) {
		t.Run("issuer", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				claims := apple.Claims{
					Email: "email",
					Claims: jwt.Claims{
						Expiry:   jwt.NewNumericDate(time.Unix(0, 0)),
						IssuedAt: jwt.NewNumericDate(time.Unix(0, 0)),
					},
				}
				err := claims.Validate("audience")
				assert.True(t, errors.Is(err, jwt.ErrInvalidIssuer))
			})

			t.Run("invalid", func(t *testing.T) {
				claims := apple.Claims{
					Email: "email",
					Claims: jwt.Claims{
						Issuer:   "invalid",
						Expiry:   jwt.NewNumericDate(time.Unix(0, 0)),
						IssuedAt: jwt.NewNumericDate(time.Unix(0, 0)),
					},
				}
				err := claims.Validate("audience")
				assert.True(t, errors.Is(err, jwt.ErrInvalidIssuer))
			})
		})

		t.Run("audience", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				claims := apple.Claims{
					Email: "email",
					Claims: jwt.Claims{
						Issuer:   apple.BaseURL,
						Expiry:   jwt.NewNumericDate(time.Unix(0, 0)),
						IssuedAt: jwt.NewNumericDate(time.Unix(0, 0)),
					},
				}
				err := claims.Validate("audience")
				assert.True(t, errors.Is(err, jwt.ErrInvalidAudience))
			})

			t.Run("invalid", func(t *testing.T) {
				claims := apple.Claims{
					Email: "email",
					Claims: jwt.Claims{
						Issuer:   apple.BaseURL,
						Audience: jwt.Audience{"invalid"},
						Expiry:   jwt.NewNumericDate(time.Unix(0, 0)),
						IssuedAt: jwt.NewNumericDate(time.Unix(0, 0)),
					},
				}
				err := claims.Validate("audience")
				assert.True(t, errors.Is(err, jwt.ErrInvalidAudience))
			})
		})

		t.Run("expired", func(t *testing.T) {
			claims := apple.Claims{
				Email: "email",
				Claims: jwt.Claims{
					Issuer:   apple.BaseURL,
					Audience: jwt.Audience{"audience"},
					// Default leeway is 1 minute
					Expiry:   jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
					IssuedAt: jwt.NewNumericDate(time.Unix(0, 0)),
				},
			}
			err := claims.Validate("audience")
			assert.True(t, errors.Is(err, jwt.ErrExpired))
		})

		t.Run("issued in the future", func(t *testing.T) {
			claims := apple.Claims{
				Email: "email",
				Claims: jwt.Claims{
					Issuer:   apple.BaseURL,
					Audience: jwt.Audience{"audience"},
					Expiry:   jwt.NewNumericDate(time.Now()),
					// Default leeway is 1 minute
					IssuedAt: jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
				},
			}
			err := claims.Validate("audience")
			assert.Error(t, err)
			assert.True(t, errors.Is(err, jwt.ErrIssuedInTheFuture))
		})
	})

	t.Run("valid", func(t *testing.T) {
		claims := apple.Claims{
			Email: "email",
			Claims: jwt.Claims{
				Issuer:   apple.BaseURL,
				Audience: jwt.Audience{"audience"},
				IssuedAt: jwt.NewNumericDate(time.Now().Add(-30 * time.Second)),
				Expiry:   jwt.NewNumericDate(time.Now().Add(30 * time.Second)),
			},
		}
		err := claims.Validate("audience")
		assert.NoError(t, err)
	})
}
