package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/models"
)

func (app *App) GetSessionRefresh(w http.ResponseWriter, req *http.Request) {
	// decode the JWT
	session, err := currentSession(app.Config, req)
	if err != nil {
		// If a session fails to decode, that's okay. Carry on.
		// TODO: log the error
	}
	if session == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// check if the session has been revoked.
	account_id, err := app.RefreshTokenStore.Find(models.RefreshToken(session.Subject))
	if err != nil {
		panic(err)
	}
	if account_id == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// refresh the refresh token
	err = app.RefreshTokenStore.Touch(models.RefreshToken(session.Subject), account_id)
	if err != nil {
		panic(err)
	}

	// generate the requested identity token
	identityToken, err := identityForSession(app.Config, session, account_id)
	if err != nil {
		panic(err)
	}

	writeData(w, http.StatusCreated, struct {
		IdToken string `json:"id_token"`
	}{
		IdToken: identityToken,
	})
}
