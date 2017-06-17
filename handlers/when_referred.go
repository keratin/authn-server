package handlers

import (
	"net/http"

	"github.com/keratin/authn/services"
)

// A specialized handler that looks like any other middleware adapter but is known to serve a
// particular purpose.
type SecurityHandler func(http.Handler) http.Handler

func RefererSecurity(domains []string) SecurityHandler {
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
				writeJson(w, http.StatusForbidden, ServiceErrors{Errors: []services.Error{services.Error{"referer", "is not a trusted host"}}})
			}
		})
	}
}
