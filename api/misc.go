package api

import (
	"net/http"
)

// A specialized handler that looks like any other middleware adapter but is known to serve a
// particular purpose.
type SecurityHandler func(http.Handler) http.Handler
