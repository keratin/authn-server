package oauth

import (
	"errors"
	"net/http"
	"time"

	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/lib/route"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/tokens/oauth"
)

func getOauth(app *api.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := app.OauthProviders[providerName]

		// require and validate a redirect URI
		redirectURI := r.FormValue("redirect_uri")
		if route.FindDomain(redirectURI, app.Config.ApplicationDomains) == nil {
			app.Reporter.ReportRequestError(errors.New("unknown redirect domain"), r)
			failsafe := app.Config.ApplicationDomains[0].URL()
			http.Redirect(w, r, failsafe.String(), http.StatusSeeOther)
			return
		}

		// fail handler
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			redirectFailure(w, r, redirectURI)
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
		stateToken, err := oauth.New(app.Config, string(nonce), redirectURI)
		if err != nil {
			fail(err)
			return
		}
		state, err := stateToken.Sign(app.Config.OAuthSigningKey)

		returnURL := app.Config.AuthNURL.String() + "/oauth/" + providerName + "/return"
		http.Redirect(w, r, provider.Config(returnURL).AuthCodeURL(state), http.StatusSeeOther)
	}
}
