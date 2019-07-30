package route

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

type matchedDomainKey int

// OriginSecurity is a SecurityHandler that will ensure a request comes from a known origin. This
// can be an effective way to mitigate CSRF attacks, which are unable to forge headers due to the
// passive nature of the attack vector.
//
// OriginSecurity will store the matching domain in the http.Request's Context. Use MatchedDomain
// to retrieve the value in later logic.
func OriginSecurity(domains []Domain, logger logrus.FieldLogger) SecurityHandler {
	var validDomains []string
	for _, d := range domains {
		validDomains = append(validDomains, d.String())
	}
	logger = logger.WithField("validDomains", validDomains)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := InferOrigin(r)
			domain := FindDomain(origin, domains)
			if domain != nil {
				ctx := r.Context()
				ctx = context.WithValue(ctx, matchedDomainKey(0), domain)

				h.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if len(origin) == 0 {
				logger.Debug("Could not infer request origin since Origin and Referer headers are both missing")
			} else {
				logger.WithField("origin", origin).Debug("Request origin is invalid")
			}
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Origin is not a trusted host."))
		})
	}
}

// MatchedDomain will retrieve from the http.Request's Context the domain that satisfied
// OriginSecurity.
func MatchedDomain(r *http.Request) *Domain {
	d, ok := r.Context().Value(matchedDomainKey(0)).(*Domain)
	if ok {
		return d
	}
	return nil
}

func InferOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin != "" {
		return origin
	}
	// If and only if the origin header is unset we can infer that this is a same-origin request
	// (i.e we trust browsers to behave this way), then we use the Referer header to discover the domain
	return r.Header.Get("Referer")
}
