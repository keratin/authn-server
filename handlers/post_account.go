package handlers

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/models"
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

	session, err := models.NewSessionJWT(
		app.RefreshTokenStore,
		app.Config,
		account.Id,
	)
	if err != nil {
		panic(err)
	}

	sessionString, err := session.Sign(jwt.SigningMethodHS256, app.Config.SessionSigningKey)
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
