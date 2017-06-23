package handlers

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/services"
	"github.com/keratin/authn/tokens"
)

type request struct {
	Username string
	Password string
}

type response struct {
	IdToken string `json:"id_token"`
}

func (app App) PostAccount(w http.ResponseWriter, req *http.Request) {
	// Create the account
	account, errors := services.AccountCreator(
		app.AccountStore,
		app.Config,
		req.FormValue("username"),
		req.FormValue("password"),
	)
	if errors != nil {
		writeErrors(w, errors)
		return
	}

	// Create the session token
	session, err := tokens.NewSessionJWT(
		app.RefreshTokenStore,
		app.Config,
		account.Id,
	)
	if err != nil {
		panic(err)
	}

	// Create the identity token
	identity, err := tokens.NewIdentityJWT(
		app.RefreshTokenStore,
		app.Config,
		session,
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	sessionString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, session).SignedString(app.Config.SessionSigningKey)
	if err != nil {
		panic(err)
	}
	sessionCookie := http.Cookie{
		Name:     app.Config.SessionCookieName,
		Value:    sessionString,
		Path:     app.Config.MountedPath,
		Secure:   app.Config.ForceSSL,
		HttpOnly: true,
	}
	http.SetCookie(w, &sessionCookie)

	// Return the identity token in the body
	identityString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, identity).SignedString(app.Config.IdentitySigningKey)
	if err != nil {
		panic(err)
	}
	writeData(w, http.StatusCreated, response{identityString})
}
