package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/server/sessions"
)

func GetOauthReturn(app *app.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := app.OauthProviders[providerName]

		// verify the state and nonce
		state, err := getState(app.Config, r)
		if err != nil {
			app.Reporter.ReportRequestError(errors.Wrap(err, "getState"), r)
			failsafe := app.Config.ApplicationDomains[0].URL()
			http.Redirect(w, r, failsafe.String(), http.StatusSeeOther)
			return
		}
		http.SetCookie(w, nonceCookie(app.Config, ""))

		// fail handler
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			redirectFailure(w, r, state.Destination)
		}

		// exchange code for tokens and user info
		returnURL := app.Config.AuthNURL.String() + "/oauth/" + providerName + "/return"
		config, err := provider.Config(returnURL)
		if err != nil {
			fail(errors.Wrap(err, "Config"))
			return
		}
		tok, err := config.Exchange(context.TODO(), r.FormValue("code"))
		if err != nil {
			fail(errors.Wrap(err, "Exchange"))
			return
		}
		providerUser, err := provider.UserInfo(tok)
		if err != nil {
			fail(errors.Wrap(err, "userInfo"))
			return
		}

		// attempt to reconcile oauth identity information into an authn account
		sessionAccountID := sessions.GetAccountID(r)
		account, err := services.IdentityReconciler(app.AccountStore, app.Config, providerName, providerUser, tok, sessionAccountID)
		if err != nil {
			fail(err)
			return
		}

		amr := []string{fmt.Sprintf("oauth:%s", providerName)}

		// identityToken is not returned in this flow. it must be imported by the frontend like a SSO session.
		sessionToken, _, err := services.SessionCreator(
			app.AccountStore, app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			account.ID, &app.Config.ApplicationDomains[0], sessions.GetRefreshToken(r), amr,
		)
		if err != nil {
			fail(errors.Wrap(err, "NewSession"))
			return
		}

		// Return the signed session in a cookie
		sessions.Set(app.Config, w, sessionToken)

		// redirect back to frontend (success or failure)
		http.Redirect(w, r, state.Destination, http.StatusSeeOther)
	}
}
