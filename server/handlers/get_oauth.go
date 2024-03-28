package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/tokens/oauth"
	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/lib/route"
)

func GetOauth(app *app.App, providerName string) http.HandlerFunc {
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
		bytes, err := lib.GenerateToken()
		if err != nil {
			fail(err)
			return
		}
		nonce := base64.StdEncoding.EncodeToString(bytes)
		http.SetCookie(w, nonceCookie(app.Config, string(nonce)))

		// save nonce and return URL into state param
		stateToken, err := oauth.New(app.Config, string(nonce), redirectURI)
		if err != nil {
			fail(err)
			return
		}
		state, err := stateToken.Sign(app.Config.OAuthSigningKey)
		if err != nil {
			fail(err)
			return
		}
		returnURL := app.Config.AuthNURL.String() + "/oauth/" + providerName + "/return"

		config, err := provider.Config(returnURL)
		if err != nil {
			fail(err)
			return
		}

		authCodeURL := config.AuthCodeURL(state, provider.AuthCodeOptions()...)

		http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
	}
}
