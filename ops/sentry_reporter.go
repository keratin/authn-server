package ops

import (
	"fmt"
	"net/http"

	raven "github.com/getsentry/raven-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

// ReportGRPCError will deliver the given error to Sentry in a background routine along with
// data relevant to the current gRPC request.
func (r *SentryReporter) ReportGRPCError(err error, info *grpc.UnaryServerInfo, req interface{}) {
	grpcErrorCode := grpc.Code(err)

	if grpcErrorCode == codes.OK {
		return
	}

	r.CaptureError(err, map[string]string{
		"server":      fmt.Sprintf("%T", info.Server),
		"grpcMethod":  info.FullMethod,
		"code":        grpcErrorCode.String(),
		"requestType": fmt.Sprintf("%T", req),
	}, nil)
}
