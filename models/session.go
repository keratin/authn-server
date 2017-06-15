package models

import (
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
)

type SessionJWT struct {
	Iss url.URL
	Sub data.RefreshToken
	Aud url.URL
	Iat time.Time
	Azp string
}

func (s *SessionJWT) claims() *jwt.MapClaims {
	return &jwt.MapClaims{
		"iss": s.Iss.String(),
		"sub": string(s.Sub),
		"aud": s.Aud.String(),
		"iat": s.Iat.Unix(),
		"azp": "", // TODO: audience
	}
}

func (s *SessionJWT) Sign(method jwt.SigningMethod, key []byte) (string, error) {
	return jwt.NewWithClaims(method, s.claims()).SignedString(key)
}

func NewSessionJWT(store data.RefreshTokenStore, cfg config.Config, account_id int) (*SessionJWT, error) {
	refreshToken, err := store.Create(account_id)
	if err != nil {
		return nil, err
	}

	session := SessionJWT{
		Iss: *cfg.AuthNURL,
		Sub: refreshToken,
		Aud: *cfg.AuthNURL,
		Iat: time.Now(),
		Azp: "", // TODO: audience
	}
	return &session, nil
}
