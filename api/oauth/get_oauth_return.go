package oauth

import (
	"context"
	"net/http"

	"github.com/keratin/authn-server/services"
	"github.com/pkg/errors"

	"github.com/keratin/authn-server/api"
)

// TODO: implement nonces
// TODO: add configuration ENV
func completeOauth(app *api.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			http.Redirect(w, r, "http://localhost:9999/TODO/FAILURE", http.StatusSeeOther)
		}

		provider := app.OauthProviders[providerName]

		// TODO: consume csrf nonce

		tok, err := provider.Config().Exchange(context.TODO(), r.FormValue("code"))
		if err != nil {
			fail(errors.Wrap(err, "Exchange"))
			return
		}
		user, err := provider.UserInfo(tok)
		if err != nil {
			fail(errors.Wrap(err, "userInfo"))
			return
		}

		account, err := app.AccountStore.FindByOauthAccount(providerName, user.ID)
		if err != nil {
			fail(errors.Wrap(err, "FindByOauthAccount"))
			return
		}

		// it's new! what to do?
		if account == nil {
			account, err = app.AccountStore.FindByUsername(user.Email)
			if err != nil {
				fail(errors.Wrap(err, "FindByUsername"))
				return
			}

			// we know this account!
			if account != nil {
				// TODO: require that a session exists and that it matches the found account.
				//       otherwise abort. we don't want an account takeover attack where someone
				//       signs up with a victim's email (unverified) and waits for the victim to
				//       connect with oauth.
				fail(errors.Wrap(err, "TODO: account exists"))
				return
			}

			// looks like a new signup!
			// TODO: if there is an existing session, then attach this oauth account to it.
			account, err = services.AccountCreator(app.AccountStore, app.Config, user.Email, "TODO: random")
			if err != nil {
				fail(errors.Wrap(err, "AccountCreator"))
				return
			}
			app.AccountStore.AddOauthAccount(account.ID, providerName, user.ID, tok.AccessToken)
		}

		// clean up any existing session
		err = api.RevokeSession(app.RefreshTokenStore, app.Config, r)
		if err != nil {
			app.Reporter.ReportRequestError(err, r)
		}

		// identityToken is not returned in this flow. it must be fetched by the frontend as a resumed session.
		sessionToken, _, err := api.NewSession(app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, account.ID, &app.Config.ApplicationDomains[0])
		if err != nil {
			fail(errors.Wrap(err, "NewSession"))
			return
		}

		// Return the signed session in a cookie
		api.SetSession(app.Config, w, sessionToken)

		// redirect back to frontend (success or failure)
		http.Redirect(w, r, "http://localhost:9999/TODO/SUCCESS", http.StatusSeeOther)
	}
}
