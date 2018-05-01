package oauth

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

func getOauth(app *api.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := app.OauthProviders[providerName]

		nonce := "TODO"
		returnURL := "TODO/oauth/" + providerName + "/return" // needs mounted path

		http.Redirect(w, r, provider.Config(returnURL).AuthCodeURL(nonce), http.StatusSeeOther)
	}
}
