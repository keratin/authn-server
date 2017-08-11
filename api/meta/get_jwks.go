package meta

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/data"
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
				KeyID:     data.RSAPublicKeyID(key.Public()),
			})
		}

		api.WriteJson(w, http.StatusOK, jose.JSONWebKeySet{Keys: keys})
	}
}
