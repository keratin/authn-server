package route

import (
	"context"
	"net/http"
	"net/url"
)

type matchedDomainKey int

// RefererSecurity is a SecurityHandler that will ensure a request comes from a known referer. This
// can be an effective way to mitigate CSRF attacks, which are unable to forge headers due to the
// passive nature of the attack vector.
//
// RefererSecurity will store the matching domain in the http.Request's Context. Use MatchedDomain
// to retrieve the value in later logic.
func RefererSecurity(domains []Domain) SecurityHandler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url, err := url.Parse(r.Header.Get("Referer"))
			if err != nil {
				panic(err)
			}

			for _, d := range domains {
				if d.Matches(url) {
					ctx := r.Context()
					ctx = context.WithValue(ctx, matchedDomainKey(0), &d)

					h.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Referer is not a trusted host."))
		})
	}
}

// MatchedDomain will retrieve from the http.Request's Context the domain that satisfied
// RefererSecurity.
func MatchedDomain(r *http.Request) *Domain {
	d, ok := r.Context().Value(matchedDomainKey(0)).(*Domain)
	if ok {
		return d
	}
	return nil
}
