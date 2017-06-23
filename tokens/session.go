package tokens

import (
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

func ParseSessionJWT(tokenStr string, cfg *config.Config) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, staticKeyFunc(cfg.SessionSigningKey))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("Could not verify JWT")
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok {
		return nil, fmt.Errorf("JWT is not a SessionClaims")
	}

	err = claims.StandardClaims.Valid()
	if err != nil {
		return nil, err
	}
	if !claims.VerifyAudience(cfg.AuthNURL.String(), true) {
		return nil, fmt.Errorf("token audience not valid")
	}
	if !claims.VerifyIssuer(cfg.AuthNURL.String(), true) {
		return nil, fmt.Errorf("token issuer not valid")
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
