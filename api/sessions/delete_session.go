package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api/sessionz"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/services"
)

func deleteSession(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := services.SessionEnder(app.RefreshTokenStore, sessionz.GetRefreshToken(r))
		if err != nil {
			app.Reporter.ReportRequestError(err, r)
		}

		sessionz.Set(app.Config, w, "")

		w.WriteHeader(http.StatusOK)
	}
}
