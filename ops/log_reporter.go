package ops

import (
	"fmt"
	"net/http"
	"time"
)

// LogReporter is an ErrorReporter that prints to a log (currently STDOUT)
type LogReporter struct{}

// ReportError reports some error information to STDOUT. The printed details are not robust.
func (r *LogReporter) ReportError(err error) {
	fmt.Printf("[%v] %v\n", time.Now(), err)
}

// ReportRequestError reports some error information to STDOUT. The printed details are not robust.
func (r *LogReporter) ReportRequestError(err error, req *http.Request) {
	fmt.Printf("[%v][%v %v] %v\n", time.Now(), req.Method, req.URL, err)
}
