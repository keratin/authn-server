package route

import "net/http"

// Unsecured is a SecurityHandler for explicitly acknowledging that a route is wide open for use.
func Unsecured() SecurityHandler {
	return func(h http.Handler) http.Handler {
		return h
	}
}
