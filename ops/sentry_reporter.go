package ops

import (
	"net/http"

	raven "github.com/getsentry/raven-go"
)

// SentryReporter is an ErrorReporter for the Sentry service (sentry.io)
type SentryReporter struct {
	*raven.Client
}

// NewSentryReporter builds a SentryReporter from a credentials string
func NewSentryReporter(dsn string) (*SentryReporter, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return nil, err
	}
	return &SentryReporter{Client: client}, nil
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
