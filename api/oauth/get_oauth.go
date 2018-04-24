package oauth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
)

func startOauth(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			http.Redirect(w, r, "http://localhost:9999/TODO/FAILURE", http.StatusSeeOther)
		}

		provider, err := getProvider(mux.Vars(r)["provider"])
		if err != nil {
			fail(err)
			return
		}

		nonce := "TODO"

		http.Redirect(w, r, provider.config().AuthCodeURL(nonce), http.StatusSeeOther)
	}
}
