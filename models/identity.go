package models

import (
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
)

type IdentityJWT struct {
	Iss      url.URL
	Sub      int
	Aud      string
	Exp      time.Time
	Iat      time.Time
	AuthTime time.Time
}

func (i *IdentityJWT) claims() *jwt.MapClaims {
	return &jwt.MapClaims{
		"iss":       i.Iss.String(),
		"sub":       i.Sub,
		"aud":       i.Aud,
		"exp":       i.Exp.Unix(),
		"iat":       i.Iat.Unix(),
		"auth_time": i.AuthTime.Unix(),
	}
}

func (i *IdentityJWT) Sign(method jwt.SigningMethod, key interface{}) (string, error) {
	return jwt.NewWithClaims(method, i.claims()).SignedString(key)
}

func NewIdentityJWT(store data.RefreshTokenStore, cfg config.Config, session *SessionJWT) (*IdentityJWT, error) {
	account_id, err := store.Find(session.Sub)
	if err != nil {
		return nil, err
	}

	identity := IdentityJWT{
		Iss:      session.Iss,
		Sub:      account_id,
		Aud:      session.Azp,
		Exp:      time.Now().Add(cfg.AccessTokenTTL),
		Iat:      time.Now(),
		AuthTime: session.Iat,
	}
	return &identity, nil
}
