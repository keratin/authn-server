package identities

import (
	"crypto/rsa"
	"strconv"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/compat"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

type Claims struct {
	AuthTime jwt.NumericDate `json:"auth_time"`
	jwt.Claims
}

func (c *Claims) Sign(rsaKey *rsa.PrivateKey) (string, error) {
	keyID, err := compat.KeyID(rsaKey.Public())
	if err != nil {
		return "", errors.Wrap(err, "KeyID")
	}

	jwk := jose.JSONWebKey{
		Key:   rsaKey,
		KeyID: keyID,
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jwk},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", errors.Wrap(err, "NewSigner")
	}
	return jwt.Signed(signer).Claims(c).CompactSerialize()
}

func New(cfg *app.Config, session *sessions.Claims, accountID int, audience string) *Claims {
	return &Claims{
		AuthTime: session.IssuedAt,
		Claims: jwt.Claims{
			Issuer:   session.Issuer,
			Subject:  strconv.Itoa(accountID),
			Audience: jwt.Audience{audience},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
}
