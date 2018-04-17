package oauth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
)

func startOauth(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider, err := getProvider(mux.Vars(r)["provider"])
		if err != nil {
			panic(err) // TODO: redirect back to frontend instead
		}

		nonce := "TODO"

		http.Redirect(w, r, provider.config().AuthCodeURL(nonce), http.StatusSeeOther)
	}
}
