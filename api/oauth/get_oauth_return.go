package oauth

import (
	"context"
	"net/http"

	"github.com/keratin/authn-server/services"

	"github.com/pkg/errors"

	"github.com/keratin/authn-server/api"
)

func getOauthReturn(app *api.App, providerName string) http.HandlerFunc {
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
		tok, err := provider.Config(returnURL).Exchange(context.TODO(), r.FormValue("code"))
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
		sessionAccountID := api.GetSessionAccountID(r)
		account, err := services.IdentityReconciler(app.AccountStore, app.Config, providerName, providerUser, tok, sessionAccountID)
		if err != nil {
			fail(err)
			return
		}

		// identityToken is not returned in this flow. it must be imported by the frontend like a SSO session.
		sessionToken, _, err := services.SessionCreator(
			app.AccountStore, app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			account.ID, &app.Config.ApplicationDomains[0], api.GetRefreshToken(r),
		)
		if err != nil {
			fail(errors.Wrap(err, "NewSession"))
			return
		}

		// Return the signed session in a cookie
		api.SetSession(app.Config, w, sessionToken)

		// redirect back to frontend (success or failure)
		http.Redirect(w, r, state.Destination, http.StatusSeeOther)
	}
}
