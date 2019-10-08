package ops

import (
	"net/http"

	"github.com/getsentry/sentry-go"
)

// SentryReporter is an ErrorReporter for the Sentry service (sentry.io)
type SentryReporter struct {
}

// NewSentryReporter builds a SentryReporter from a credentials string
func NewSentryReporter(dsn string) (*SentryReporter, error) {
	err := sentry.Init(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		return nil, err
	}
	return &SentryReporter{}, nil
}

// ReportError will deliver the given error to Sentry in a background routine.
func (r *SentryReporter) ReportError(err error) {
	sentry.CaptureException(err)
}

// ReportRequestError will deliver the given error to Sentry in a background routine along with
// data relevant to the current http.Request.
//
// NOTE: POST data is never reported to Sentry, so passwords remain private.
func (r *SentryReporter) ReportRequestError(err error, req *http.Request) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetRequest(sentry.Request{}.FromHTTPRequest(req))
		sentry.CaptureException(err)
	})
}
