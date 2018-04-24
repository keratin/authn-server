package oauth

import (
	"context"
	"net/http"

	"github.com/keratin/authn-server/services"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

// TODO: create database table for service identities (service name, service id, access token, refresh token)
// TODO: implement AccountStore updates
// TODO: implement nonces
// TODO: add configuration ENV
// TODO: delete oauth accounts when deleting account
// TODO: is oauth2.NoContext deprecated?
//
// TODO: how does the app get a token it can use for api calls with the scopes it has defined? do we
//       need to save the current oauth token to the database? probably in whatever table contains
//       the provider identities per account.
func completeOauth(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			http.Redirect(w, r, "http://localhost:9999/TODO/FAILURE", http.StatusSeeOther)
		}

		providerName := mux.Vars(r)["provider"]
		provider, err := getProvider(providerName)
		if err != nil {
			fail(err)
			return
		}

		// TODO: consume csrf nonce

		tok, err := provider.config().Exchange(context.TODO(), r.FormValue("code"))
		if err != nil {
			fail(err)
			return
		}
		user, err := provider.userInfo(tok)
		if err != nil {
			fail(err)
			return
		}

		account, err := app.AccountStore.FindByOauthAccount(providerName, user.id)
		if err != nil {
			fail(err)
			return
		}

		// it's new! what to do?
		if account != nil {
			account, err = app.AccountStore.FindByUsername(user.email)
			if err != nil {
				fail(err)
				return
			}

			// we know this account!
			if account != nil {
				// TODO: require that a session exists and that it matches the found account.
				//       otherwise abort. we don't want an account takeover attack where someone
				//       signs up with a victim's email (unverified) and waits for the victim to
				//       connect with oauth.
				fail(err)
				return
			}

			// looks like a new signup!
			// TODO: but wait, what if there's an existing session? do we add an oauth account to it?
			account, err = services.AccountCreator(app.AccountStore, app.Config, user.email, "")
			if err != nil {
				fail(err)
				return
			}
			app.AccountStore.AddOauthAccount(account.ID, providerName, user.id, tok.AccessToken)
		}

		// clean up any existing session
		err = api.RevokeSession(app.RefreshTokenStore, app.Config, r)
		if err != nil {
			app.Reporter.ReportRequestError(err, r)
		}

		// identityToken is not returned in this flow. it must be fetched by the frontend as a resumed session.
		sessionToken, _, err := api.NewSession(app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, account.ID, route.MatchedDomain(r))
		if err != nil {
			panic(err)
		}

		// Return the signed session in a cookie
		api.SetSession(app.Config, w, sessionToken)

		// redirect back to frontend (success or failure)
		http.Redirect(w, r, "http://localhost:9999/TODO/SUCCESS", http.StatusSeeOther)
	}
}
