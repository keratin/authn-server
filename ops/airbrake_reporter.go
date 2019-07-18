package ops

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/airbrake/gobrake"
)

// AirbrakeReporter is an ErrorReporter for the Airbrake service (airbrake.io)
type AirbrakeReporter struct {
	*gobrake.Notifier
}

// NewAirbrakeReporter builds an AirbrakeReporter from a credentials string. The credentials string
// should be in the pattern $PROJECT_ID:$PROJECT_KEY (aka username:password).
func NewAirbrakeReporter(credentials string) (*AirbrakeReporter, error) {
	bits := strings.SplitN(credentials, ":", 2)
	projectID, err := strconv.Atoi(bits[0])
	if err != nil {
		return nil, err
	}
	projectKey := bits[1]
	client := gobrake.NewNotifier(int64(projectID), projectKey)
	return &AirbrakeReporter{Notifier: client}, nil
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
