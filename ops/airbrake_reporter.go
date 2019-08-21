package ops

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/airbrake/gobrake"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

// ReportRequestError will deliver the given error to Airbrake in a background routine along with
// data relevant to the current http.Request.
//
// NOTE: POST data is never reported to Airbrake, so passwords remain private.
func (r *AirbrakeReporter) ReportRequestError(err error, req *http.Request) {
	r.Notify(err, req)
}

// ReportGRPCError will deliver the given error to Airbrake in a background routine along with
// data relevant to the current gRPC request.
func (r *AirbrakeReporter) ReportGRPCError(err error, info *grpc.UnaryServerInfo, req interface{}) {
	grpcErrorCode := grpc.Code(err)

	if grpcErrorCode == codes.OK {
		return
	}

	r.Notify(&gobrake.Notice{
		Errors: []gobrake.Error{
			gobrake.Error{
				Type:    fmt.Sprintf("%T", err),
				Message: err.Error(),
			},
		},
		Context: map[string]interface{}{
			"server":      fmt.Sprintf("%T", info.Server),
			"grpcMethod":  info.FullMethod,
			"code":        grpcErrorCode.String(),
			"requestType": fmt.Sprintf("%T", req),
		},
	}, nil)
}
