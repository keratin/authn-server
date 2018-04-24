package oauth

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

func startOauth(app *api.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := app.OauthProviders[providerName]

		nonce := "TODO"

		http.Redirect(w, r, provider.Config().AuthCodeURL(nonce), http.StatusSeeOther)
	}
}
