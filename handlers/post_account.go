package handlers

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/services"
)

type request struct {
	Username string
	Password string
}

type response struct {
	IdToken string `json:"id_token"`
}

func (app App) PostAccount(w http.ResponseWriter, req *http.Request) {
	account, errors := services.AccountCreator(
		app.AccountStore,
		&app.Config,
		req.FormValue("username"),
		req.FormValue("password"),
	)
	if errors != nil {
		writeErrors(w, errors)
		return
	}

	refreshToken, err := app.RefreshTokenStore.Create(account.Id)
	if err != nil {
		panic(err)
	}

	sessionToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": app.Config.AuthNURL.String(),
		"sub": refreshToken,
		"aud": app.Config.AuthNURL.String(),
		"iat": time.Now().Unix(),
		"azp": "",
	})
	sessionString, err := sessionToken.SignedString(app.Config.SessionSigningKey)
	if err != nil {
		panic(err)
	}
	sessionCookie := http.Cookie{
		Name:     "authn",
		Value:    sessionString,
		Path:     app.Config.MountedPath,
		Secure:   app.Config.ForceSSL,
		HttpOnly: true,
	}

	accessToken := "j.w.t"

	w.WriteHeader(http.StatusCreated)
	http.SetCookie(w, &sessionCookie)
	writeData(w, response{accessToken})
}
