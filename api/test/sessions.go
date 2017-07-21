package test

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
)

func CreateSession(tokenStore data.RefreshTokenStore, cfg *config.Config, accountId int) *http.Cookie {
	sessionToken, err := sessions.New(tokenStore, cfg, accountId)
	if err != nil {
		panic(err)
	}

	sessionString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, sessionToken).SignedString(cfg.SessionSigningKey)
	if err != nil {
		panic(err)
	}

	return &http.Cookie{
		Name:  cfg.SessionCookieName,
		Value: sessionString,
	}
}

func RevokeSession(store data.RefreshTokenStore, cfg *config.Config, session *http.Cookie) {
	claims, err := sessions.Parse(session.Value, cfg)
	if err != nil {
		panic(err)
	}
	err = store.Revoke(models.RefreshToken(claims.Subject))
	if err != nil {
		panic(err)
	}
}
