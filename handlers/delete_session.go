package handlers

import "net/http"

func DeleteSession(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := revokeSession(app.RefreshTokenStore, app.Config, req)
		if err != nil {
			// TODO: alert but continue
		}

		setSession(app.Config, w, "")

		w.WriteHeader(http.StatusOK)
	}
}
