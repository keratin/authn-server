package tokens

import (
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
	"github.com/keratin/authn/models"
)

type IdentityClaims struct {
	AuthTime int64 `json:"auth_time"`
	jwt.StandardClaims
}

func (c *IdentityClaims) Valid() error {
	return nil
}

func NewIdentityJWT(store data.RefreshTokenStore, cfg *config.Config, session *SessionClaims) (*IdentityClaims, error) {
	account_id, err := store.Find(models.RefreshToken(session.Subject))
	if err != nil {
		return nil, err
	}

	return &IdentityClaims{
		AuthTime: session.IssuedAt,
		StandardClaims: jwt.StandardClaims{
			Issuer:    session.Issuer,
			Subject:   strconv.Itoa(account_id),
			Audience:  session.Azp,
			ExpiresAt: time.Now().Add(cfg.AccessTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}, nil
}
