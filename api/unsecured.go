package api

import "net/http"

func Unsecured() SecurityHandler {
	return func(h http.Handler) http.Handler {
		return h
	}
}
