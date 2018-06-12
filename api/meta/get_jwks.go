package meta

import (
	"net/http"

	"github.com/keratin/authn-server/lib/compat"

	"github.com/keratin/authn-server/api"
	jose "gopkg.in/square/go-jose.v2"
)

func getJWKs(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys := []jose.JSONWebKey{}
		for _, key := range app.KeyStore.Keys() {
			keyID, err := compat.KeyID(key.Public())
			if err != nil {
				app.Reporter.ReportRequestError(err, r)
			} else {
				keys = append(keys, jose.JSONWebKey{
					Key:       key.Public(),
					Use:       "sig",
					Algorithm: "RS256",
					KeyID:     keyID,
				})
			}
		}

		api.WriteJSON(w, http.StatusOK, jose.JSONWebKeySet{Keys: keys})
	}
}
