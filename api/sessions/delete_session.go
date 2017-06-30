package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

func DeleteSession(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := api.RevokeSession(app.RefreshTokenStore, app.Config, req)
		if err != nil {
			// TODO: alert but continue
		}

		api.SetSession(app.Config, w, "")

		w.WriteHeader(http.StatusOK)
	}
}
