package identities

import (
	"strconv"
	"time"

	"github.com/keratin/authn-server/app/data/private"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// CLaims represent the JWT claims known to authn.
// Note that this type is embedded in the go client when changing:
// https://github.com/keratin/authn-go/blob/master/authn/claims.go
type Claims struct {
	AuthTime  *jwt.NumericDate `json:"auth_time"`
	SessionID string           `json:"sid"`
	jwt.Claims
}

func (c *Claims) Sign(key *private.Key) (string, error) {
	jwk := jose.JSONWebKey{
		Key:   key.PrivateKey,
		KeyID: key.JWK.KeyID,
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
		AuthTime:  session.IssuedAt,
		SessionID: session.SessionID,
		Claims: jwt.Claims{
			Issuer:   session.Issuer,
			Subject:  strconv.Itoa(accountID),
			Audience: jwt.Audience{audience},
			Expiry:   jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
}
