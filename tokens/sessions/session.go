package sessions

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
)

type Claims struct {
	Azp string `json:"azp"`
	jwt.StandardClaims
}

func (c *Claims) Sign(hmac_key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(hmac_key)
}

func Parse(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, staticKeyFunc(cfg.SessionSigningKey))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("Could not verify JWT")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("JWT is not a Claims")
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

func New(store data.RefreshTokenStore, cfg *config.Config, account_id int) (*Claims, error) {
	refreshToken, err := store.Create(account_id)
	if err != nil {
		return nil, err
	}

	return &Claims{
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
