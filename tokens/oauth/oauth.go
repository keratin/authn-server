package oauth

import (
	"fmt"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

const scope = "oauth"

// Claims is a JWT intended to be used as the state param in an OAuth exchange. It wraps a nonce and
// a return URL in a signed, tamper-proof string.
// See: https://tools.ietf.org/html/draft-bradley-oauth-jwt-encoded-state-00
type Claims struct {
	Scope                    string `json:"scope"`
	RequestForgeryProtection string `json:"rfp"`
	Destination              string `json:"dst"`
	jwt.Claims
}

// Sign converts the claims into a serialized string, signed with HMAC.
func (c *Claims) Sign(hmacKey []byte) (string, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: hmacKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", errors.Wrap(err, "NewSigner")
	}
	return jwt.Signed(signer).Claims(c).CompactSerialize()
}

// Parse will deserialize a string into Claims if and only if the claims pass all validations. In
// this case the token must contain a nonce already known from a different channel (like a cookie).
func Parse(tokenStr string, cfg *app.Config, nonce string) (*Claims, error) {
	token, err := jwt.ParseSigned(tokenStr)
	if err != nil {
		return nil, errors.Wrap(err, "ParseSigned")
	}

	claims := Claims{}
	err = token.Claims(cfg.OAuthSigningKey, &claims)
	if err != nil {
		return nil, errors.Wrap(err, "Claims")
	}

	err = claims.Claims.Validate(jwt.Expected{
		Audience: jwt.Audience{cfg.AuthNURL.String()},
		Issuer:   cfg.AuthNURL.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Validate")
	}
	if claims.Scope != scope {
		return nil, fmt.Errorf("token scope not valid")
	}
	if claims.RequestForgeryProtection != nonce {
		return nil, fmt.Errorf("nonce does not match")
	}

	return &claims, nil
}

// New creates Claims for a JWT suitable as a state parameter during an OAuth flow.
func New(cfg *app.Config, nonce string, destination string) (*Claims, error) {
	return &Claims{
		Scope: scope,
		RequestForgeryProtection: nonce,
		Destination:              destination,
		Claims: jwt.Claims{
			Issuer:   cfg.AuthNURL.String(),
			Audience: jwt.Audience{cfg.AuthNURL.String()},
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}, nil
}
