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

// PanicHandler returns a http.Handler that will recover any panics and print some information to
// STDOUT. The printed details are not robust. If a panic is caught, the handler will return HTTP
// 500.
func (r *LogReporter) PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Printf("[%v] %v", time.Now(), err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, req)
	})
}
