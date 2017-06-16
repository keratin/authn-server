package handlers

import (
	"net/http"

	"github.com/keratin/authn/services"
)

type adapter func(http.Handler) http.Handler

func WhenReferred(domains []string) adapter {
	// optimize for lookups
	domainMap := make(map[string]bool)
	for _, i := range domains {
		domainMap[i] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if domainMap[r.Header.Get("REFERER")] {
				h.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
				writeJson(w, ServiceErrors{Errors: []services.Error{services.Error{"referer", "is not a trusted host"}}})
			}
		})
	}
}
