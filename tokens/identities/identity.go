package identities

import (
	"crypto/rsa"
	"strconv"
	"time"

	"github.com/keratin/authn-server/compat"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/tokens/sessions"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

type Claims struct {
	AuthTime jwt.NumericDate `json:"auth_time"`
	jwt.Claims
}

func (c *Claims) Sign(rsaKey *rsa.PrivateKey) (string, error) {
	jwk := jose.JSONWebKey{
		Key:   rsaKey,
		KeyID: compat.KeyID(rsaKey.Public()),
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jwk},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return "", err
	}
	return jwt.Signed(signer).Claims(c).CompactSerialize()
}

func New(cfg *config.Config, session *sessions.Claims, accountId int) *Claims {
	return &Claims{
		AuthTime: session.IssuedAt,
		Claims: jwt.Claims{
			Issuer:   session.Issuer,
			Subject:  strconv.Itoa(accountId),
			Audience: jwt.Audience{session.Azp},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
}
