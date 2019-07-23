package ops

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// LogReporter is an ErrorReporter that prints to the log (likely STDOUT)
type LogReporter struct {
	logrus.FieldLogger
}

// ReportError logs error information. The printed details are not robust.
func (r *LogReporter) ReportError(err error) {
	r.Error(err)
}

// ReportRequestError logs error information. The printed details are not robust.
func (r *LogReporter) ReportRequestError(err error, req *http.Request) {
	r.WithFields(logrus.Fields{"method": req.Method, "URL": req.URL}).Error(err)
}
