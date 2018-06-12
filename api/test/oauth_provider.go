package test

import (
	"encoding/json"
	"net/http"
)

// ProviderApp returns a a HandlerFunc that can be served as a test double for a fake OAuth provider
func ProviderApp() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			j, _ := json.Marshal(map[string]interface{}{
				"access_token":  r.FormValue("code"),
				"refresh_token": "REFRESHTOKEN",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
