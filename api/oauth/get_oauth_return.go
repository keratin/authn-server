package oauth

import (
	"context"
	"net/http"

	"github.com/keratin/authn-server/models"

	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/services"
	"github.com/pkg/errors"

	"github.com/keratin/authn-server/api"
)

// TODO: implement nonce or state check
// TODO: add return URL configuration
// TODO: add configuration ENVs
func completeOauth(app *api.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			http.Redirect(w, r, "http://localhost:9999/TODO/FAILURE", http.StatusSeeOther)
		}

		succeed := func(account *models.Account) {
			// clean up any existing session
			err := api.RevokeSession(app.RefreshTokenStore, app.Config, r)
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

		provider := app.OauthProviders[providerName]

		// TODO: consume csrf nonce

		tok, err := provider.Config().Exchange(context.TODO(), r.FormValue("code"))
		if err != nil {
			fail(errors.Wrap(err, "Exchange"))
			return
		}
		providerUser, err := provider.UserInfo(tok)
		if err != nil {
			fail(errors.Wrap(err, "userInfo"))
			return
		}

		// LOGGING IN
		// Require a previously linked account.
		linkedAccount, err := app.AccountStore.FindByOauthAccount(providerName, providerUser.ID)
		if err != nil {
			fail(errors.Wrap(err, "FindByOauthAccount"))
			return
		}
		if linkedAccount != nil {
			succeed(linkedAccount)
			return
		}

		// CONNECTING ACCOUNTS
		// Require a session with an account that has no other identity from this provider.
		sessionAccountID := api.GetSessionAccountID(r)
		if sessionAccountID != 0 {
			sessionAccountIdentities, err := app.AccountStore.GetOauthAccounts(sessionAccountID)
			if err != nil {
				fail(errors.Wrap(err, "GetOauthAccounts"))
				return
			}
			sessionAccountHasExistingIdentity := false
			for _, i := range sessionAccountIdentities {
				if i.Provider == providerName {
					sessionAccountHasExistingIdentity = true
				}
			}
			if !sessionAccountHasExistingIdentity {
				err = app.AccountStore.AddOauthAccount(sessionAccountID, providerName, providerUser.ID, tok.AccessToken)
				if err != nil {
					fail(errors.Wrap(err, "AddOauthAccount"))
					return
				}
				sessionAccount, err := app.AccountStore.Find(sessionAccountID)
				if err != nil {
					fail(errors.Wrap(err, "Find"))
					return
				}
				succeed(sessionAccount)
				return
			}
		}

		// SIGNING UP
		// Require a unique email (username)
		rand, err := lib.GenerateToken()
		if err != nil {
			fail(errors.Wrap(err, "GenerateToken"))
			return
		}
		newAccount, err := services.AccountCreator(app.AccountStore, app.Config, providerUser.Email, string(rand))
		if err != nil {
			fail(errors.Wrap(err, "AccountCreator"))
			return
		}
		app.AccountStore.AddOauthAccount(newAccount.ID, providerName, providerUser.ID, tok.AccessToken)
		succeed(newAccount)
		return
	}
}
