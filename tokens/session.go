package tokens

import (
	"errors"
	"fmt"
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

func ParseSessionJWT(tokenStr string, sessionSigningKey []byte) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, staticKeyFunc(sessionSigningKey))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Could not verify JWT")
	}
	return claims, nil
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

func staticKeyFunc(key []byte) func(*jwt.Token) (interface{}, error) {
	return func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	}
}
