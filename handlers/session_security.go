package handlers

import (
	"context"
	"net/http"

	"github.com/keratin/authn-server/models"
)

func SessionSecurity(app *App) SecurityHandler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// decode the JWT
			session, err := currentSession(app.Config, r)
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

			// store some useful things. any endpoint that requires a session will
			// want this information.
			ctx := r.Context()
			ctx = context.WithValue(ctx, SessionKey, session)
			ctx = context.WithValue(ctx, AccountIDKey, account_id)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
