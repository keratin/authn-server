package ops

import (
	"log"
	"net/http"
	"time"
)

// LogReporter is an ErrorReporter that prints to the log (likely STDOUT)
type LogReporter struct{}

// ReportError logs error information. The printed details are not robust.
func (r *LogReporter) ReportError(err error) {
	log.Printf("[%v] %v\n", time.Now(), err)
}

// ReportRequestError logs error information. The printed details are not robust.
func (r *LogReporter) ReportRequestError(err error, req *http.Request) {
	log.Printf("[%v][%v %v] %v\n", time.Now(), req.Method, req.URL, err)
}
