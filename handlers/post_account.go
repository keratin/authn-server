package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/services"
)

func (app *App) PostAccount(w http.ResponseWriter, req *http.Request) {
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

	err := revokeSession(app.RefreshTokenStore, app.Config, req)
	if err != nil {
		// TODO: alert but continue
	}

	sessionToken, identityToken, err := establishSession(app.RefreshTokenStore, app.Config, account.Id)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	setSession(app.Config, w, sessionToken)

	// Return the signed identity token in the body
	writeData(w, http.StatusCreated, struct {
		IdToken string `json:"id_token"`
	}{
		IdToken: identityToken,
	})
}
