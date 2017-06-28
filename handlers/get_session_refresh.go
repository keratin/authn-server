package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
)

func (app *App) GetSessionRefresh(w http.ResponseWriter, req *http.Request) {
	session := req.Context().Value(SessionKey).(*sessions.Claims)
	account_id := req.Context().Value(AccountIDKey).(int)

	err := app.RefreshTokenStore.Touch(models.RefreshToken(session.Subject), account_id)
	if err != nil {
		panic(err)
	}

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
