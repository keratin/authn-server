package ops

import (
	"net/http"

	raven "github.com/getsentry/raven-go"
)

// SentryReporter is an ErrorReporter for the Sentry service (sentry.io)
type SentryReporter struct {
	*raven.Client
}

// ReportError will deliver the given error to Sentry in a background routine.
func (r *SentryReporter) ReportError(err error) {
	r.CaptureError(err, map[string]string{})
}

// ReportRequestError will deliver the given error to Sentry in a background routine along with
// data relevant to the current http.Request.
//
// NOTE: POST data is never reported to Sentry, so passwords remain private.
func (r *SentryReporter) ReportRequestError(err error, req *http.Request) {
	r.CaptureError(err, map[string]string{}, raven.NewHttp(req))
}
