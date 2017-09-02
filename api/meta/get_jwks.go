package meta

import (
	"net/http"

	"github.com/keratin/authn-server/compat"

	"github.com/keratin/authn-server/api"
	jose "github.com/square/go-jose"
)

func getJWKs(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys := []jose.JSONWebKey{}
		for _, key := range app.KeyStore.Keys() {
			keys = append(keys, jose.JSONWebKey{
				Key:       key.Public(),
				Use:       "sig",
				Algorithm: "RS256",
				KeyID:     compat.KeyID(key.Public()),
			})
		}

		api.WriteJson(w, http.StatusOK, jose.JSONWebKeySet{Keys: keys})
	}
}
