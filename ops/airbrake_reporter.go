package ops

import (
	"net/http"

	"github.com/airbrake/gobrake"
)

// AirbrakeReporter is an ErrorReporter for the Airbrake service (airbrake.io)
type AirbrakeReporter struct {
	*gobrake.Notifier
}

// ReportError will deliver the given error to Airbrake in a background routine
func (r *AirbrakeReporter) ReportError(err error) {
	r.Notify(err, nil)
}

// ReportRequestError will deliver the given error to Sentry in a background routine along with
// data relevant to the current http.Request.
//
// NOTE: POST data is never reported to Airbrake, so passwords remain private.
func (r *AirbrakeReporter) ReportRequestError(err error, req *http.Request) {
	r.Notify(err, req)
}
