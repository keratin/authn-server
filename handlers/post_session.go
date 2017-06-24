package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/services"
)

func (app *App) PostSession(w http.ResponseWriter, req *http.Request) {
	// Check the password
	account, errors := services.CredentialsVerifier(
		app.AccountStore,
		app.Config,
		req.FormValue("username"),
		req.FormValue("password"),
	)
	if errors != nil {
		writeErrors(w, errors)
		return
	}

	err := revokeSession(app.RefreshTokenStore, app.Config, req)
	if err != nil {
		// TODO: alert but continue
	}

	sessionToken, identityToken, err := establishSession(app.RefreshTokenStore, app.Config, account.Id)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     app.Config.SessionCookieName,
		Value:    sessionToken,
		Path:     app.Config.MountedPath,
		Secure:   app.Config.ForceSSL,
		HttpOnly: true,
	})

	// Return the signed identity token in the body
	writeData(w, http.StatusCreated, struct {
		IdToken string `json:"id_token"`
	}{
		IdToken: identityToken,
	})
}
