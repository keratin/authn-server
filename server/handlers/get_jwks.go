package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"gopkg.in/square/go-jose.v2"
)

func GetJWKs(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var keys []jose.JSONWebKey
		for _, key := range app.KeyStore.Keys() {
			keys = append(keys, key.JWK)
		}

		WriteJSON(w, http.StatusOK, jose.JSONWebKeySet{Keys: keys})
	}
}
