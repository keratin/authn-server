package tokens

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
)

type SessionClaims struct {
	Azp string `json:"azp"`
	jwt.StandardClaims
}

func (c SessionClaims) Valid() error {
	return nil
}

func NewSessionJWT(store data.RefreshTokenStore, cfg *config.Config, account_id int) (*SessionClaims, error) {
	refreshToken, err := store.Create(account_id)
	if err != nil {
		return nil, err
	}

	return &SessionClaims{
		Azp: "", // TODO: audience
		StandardClaims: jwt.StandardClaims{
			Issuer:   cfg.AuthNURL.String(),
			Subject:  string(refreshToken),
			Audience: cfg.AuthNURL.String(),
			IssuedAt: time.Now().Unix(),
		},
	}, nil
}
