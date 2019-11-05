package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/sessions"
)

//GetTOTP begins the TOTP onboarding process
func GetTOTP(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		img, err := services.TOTPCreator(app.AccountStore, app.TOTPCache, accountID, route.MatchedDomain(r))
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(img.Bytes())))
		if _, err := w.Write(img.Bytes()); err != nil {
			log.Fatal(err)
		}
	}
}
