package identities

import (
	"crypto/rsa"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/tokens/sessions"
)

type Claims struct {
	AuthTime int64 `json:"auth_time"`
	jwt.StandardClaims
}

func (c *Claims) Sign(rsa_key *rsa.PrivateKey) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(rsa_key)
}

func New(cfg *config.Config, session *sessions.Claims, accountId int) *Claims {
	return &Claims{
		AuthTime: session.IssuedAt,
		StandardClaims: jwt.StandardClaims{
			Issuer:    session.Issuer,
			Subject:   strconv.Itoa(accountId),
			Audience:  session.Azp,
			ExpiresAt: time.Now().Add(cfg.AccessTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
}
