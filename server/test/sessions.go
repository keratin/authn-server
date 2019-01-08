package test

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/tokens/sessions"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

func CreateSession(tokenStore data.RefreshTokenStore, cfg *app.Config, accountID int) *http.Cookie {
	sessionToken, err := sessions.New(tokenStore, cfg, accountID, cfg.ApplicationDomains[0].String())
	if err != nil {
		panic(err)
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: cfg.SessionSigningKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		panic(err)
	}
	sessionString, err := jwt.Signed(signer).Claims(sessionToken).CompactSerialize()
	if err != nil {
		panic(err)
	}

	return &http.Cookie{
		Name:  cfg.SessionCookieName,
		Value: sessionString,
	}
}

func RevokeSession(store data.RefreshTokenStore, cfg *app.Config, session *http.Cookie) {
	claims, err := sessions.Parse(session.Value, cfg)
	if err != nil {
		panic(err)
	}
	err = store.Revoke(models.RefreshToken(claims.Subject))
	if err != nil {
		panic(err)
	}
}
