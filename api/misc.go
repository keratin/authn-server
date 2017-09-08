package api

import (
	"net/http"
)

// SecurityHandler is a kind of middleware that satisfies the security criteria of AuthN's routing
type SecurityHandler func(http.Handler) http.Handler
