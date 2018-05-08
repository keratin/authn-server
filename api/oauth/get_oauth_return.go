package oauth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/oauth"

	"github.com/keratin/authn-server/lib"
	"github.com/keratin/authn-server/services"
	"github.com/pkg/errors"

	"github.com/keratin/authn-server/api"
)

func getState(cfg *config.Config, r *http.Request) (*oauth.Claims, error) {
	nonce, err := r.Cookie(cfg.OAuthCookieName)
	if err != nil {
		return nil, errors.Wrap(err, "Cookie")
	}
	state, err := oauth.Parse(r.FormValue("state"), cfg, nonce.Value)
	if err != nil {
		return nil, errors.Wrap(err, "Parse")
	}
	return state, err
}

func redirectFailure(w http.ResponseWriter, r *http.Request, destination string) {
	url, _ := url.Parse(destination)
	query := url.Query()
	query.Add("status", "failed")
	url.RawQuery = query.Encode()
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

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

		// fail handler
		fail := func(err error) {
			app.Reporter.ReportRequestError(err, r)
			redirectFailure(w, r, state.Destination)
		}

		// success handler
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
			http.Redirect(w, r, state.Destination, http.StatusSeeOther)
		}

		// exchange code for tokens and user info
		tok, err := provider.Config("TODO").Exchange(context.TODO(), r.FormValue("code"))
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
			if linkedAccount.Locked {
				fail(errors.New("account locked"))
				return
			}
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
