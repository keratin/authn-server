package ops

import (
	"net/http"
)

// ErrorReporter is a thing that exports details about errors and panics to another service. Care
// must be taken by each implementation to ensure that passwords are not leaked.
type ErrorReporter interface {
	ReportError(err error)
	ReportRequestError(err error, r *http.Request)
	PanicHandler(http http.Handler) http.Handler
}
