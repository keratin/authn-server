package handlers

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/sessions"
)

// GetTOTP begins the TOTP onboarding process
func GetTOTP(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		totpKey, err := services.TOTPCreator(app.AccountStore, app.TOTPCache, accountID, route.MatchedDomain(r))
		if err != nil {
			panic(err)
		}

		//Convert to png
		var pngImg bytes.Buffer
		img, err := totpKey.Image(200, 200)
		if err != nil {
			panic(err)
		}
		err = png.Encode(&pngImg, img)
		if err != nil {
			panic(err)
		}

		WriteData(w, http.StatusOK, map[string]string{
			"secret": totpKey.Secret(),
			"url":    totpKey.URL(),
			"png":    base64.StdEncoding.EncodeToString(pngImg.Bytes()),
		})
	}
}
