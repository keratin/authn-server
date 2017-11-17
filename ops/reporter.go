package ops

import (
	"fmt"
	"net/http"
	"time"

	raven "github.com/getsentry/raven-go"
)

// ErrorReporter is a thing that exports details about errors and panics to another service. Care
// must be taken by each implementation to ensure that passwords are not leaked.
type ErrorReporter interface {
	ReportError(err error)
	ReportRequestError(err error, r *http.Request)
	PanicHandler(http http.Handler) http.Handler
}

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

// PanicHandler returns a http.Handler that will recover any panics and deliver them to Sentry
// in a background routine. If a panic is caught, the handler will return HTTP 500.
//
// NOTE: POST data is never reported to Sentry, so passwords remain private.
func (r *SentryReporter) PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err, _ := r.CapturePanic(
			func() { next.ServeHTTP(w, req) },
			map[string]string{},
			raven.NewHttp(req),
		)
		if err != nil {
			// TODO: log the sentryID for forensics
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
