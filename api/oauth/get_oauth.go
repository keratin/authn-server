package oauth

import (
	"net/http"
	"time"

	"github.com/keratin/authn-server/lib"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/tokens/oauth"
)

func getOauth(app *api.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := app.OauthProviders[providerName]
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			http.Redirect(w, r, "http://localhost:9999/TODO/FAILURE", http.StatusSeeOther)
		}

		// set nonce in a secured cookie
		nonce, err := lib.GenerateToken()
		if err != nil {
			fail(err)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     app.Config.OAuthCookieName,
			Value:    string(nonce),
			Path:     app.Config.MountedPath,
			Secure:   app.Config.ForceSSL,
			HttpOnly: true,
			MaxAge:   int(time.Hour.Seconds()),
		})

		// save nonce and return URL into state param
		stateToken, err := oauth.New(app.Config, string(nonce), "http://localhost:9999/TODO/SUCCESS")
		if err != nil {
			fail(err)
			return
		}
		state, err := stateToken.Sign(app.Config.OAuthSigningKey)

		returnURL := "TODO/oauth/" + providerName + "/return" // needs mounted path
		http.Redirect(w, r, provider.Config(returnURL).AuthCodeURL(state), http.StatusSeeOther)
	}
}
