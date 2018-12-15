package handlers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/tokens/oauth"
	"github.com/pkg/errors"
)

// nonceCookie creates or deletes a cookie containing val (the nonce)
func nonceCookie(cfg *app.Config, val string) *http.Cookie {
	var maxAge int
	if val == "" {
		maxAge = -1
	} else {
		maxAge = int(time.Hour.Seconds())
	}

	return &http.Cookie{
		Name:     cfg.OAuthCookieName,
		Value:    val,
		Path:     cfg.MountedPath,
		Secure:   cfg.ForceSSL,
		HttpOnly: true,
		MaxAge:   maxAge,
	}
}

// getState returns a verified state token using the nonce cookie
func getState(cfg *app.Config, r *http.Request) (*oauth.Claims, error) {
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

// redirectFailure is a redirect with status=failed added to the destination
func redirectFailure(w http.ResponseWriter, r *http.Request, destination string) {
	url, _ := url.Parse(destination)
	query := url.Query()
	query.Add("status", "failed")
	url.RawQuery = query.Encode()
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}
