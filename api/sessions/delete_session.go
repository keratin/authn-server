package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

func deleteSession(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := api.RevokeSession(app.RefreshTokenStore, app.Config, r)
		if err != nil {
			app.Reporter.ReportError(err)
		}

		api.SetSession(app.Config, w, "")

		w.WriteHeader(http.StatusOK)
	}
}
