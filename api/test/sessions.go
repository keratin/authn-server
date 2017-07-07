package test

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/tokens/sessions"
)

func CreateSession(tokenStore data.RefreshTokenStore, cfg *config.Config, account_id int) *http.Cookie {
	sessionToken, err := sessions.New(tokenStore, cfg, account_id)
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
