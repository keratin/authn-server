package ops

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// ErrorReporterType exists to provide context for plain strings in configuration
type ErrorReporterType int

// all known types of ErrorReporter
const (
	Log ErrorReporterType = iota
	Sentry
	Airbrake
)

// ErrorReporter is a thing that exports details about errors and panics to another service. Care
// must be taken by each implementation to ensure that passwords are not leaked.
type ErrorReporter interface {
	ReportError(err error)
	ReportRequestError(err error, r *http.Request)
	ReportGRPCError(err error, info *grpc.UnaryServerInfo, req interface{})
}

// NewErrorReporter will instantiate an ErrorReporter for a known type
func NewErrorReporter(credentials string, t ErrorReporterType, logger logrus.FieldLogger) (ErrorReporter, error) {
	switch t {
	case Sentry:
		return NewSentryReporter(credentials)
	case Airbrake:
		return NewAirbrakeReporter(credentials)
	default:
		return &LogReporter{
			FieldLogger: logger.WithField("scope", "NewErrorReporter"),
		}, nil
	}
}

// PanicHandler returns a http.Handler that will recover any panics and report them as request
// errors. If a panic is caught, the handler will return HTTP 500.
func PanicHandler(r ErrorReporter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			val := recover()
			switch err := val.(type) {
			case nil:
				return
			case error:
				r.ReportRequestError(err, req)
				w.WriteHeader(http.StatusInternalServerError)
			default:
				r.ReportRequestError(errors.New(fmt.Sprint(err)), req)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, req)
	})
}

// GRPCRecoveryInterceptor return an interceptor that will recover any panics and report them as request
// errors. If a panic is caught, the interceptor will return HTTP 500.
func GRPCRecoveryInterceptor(r ErrorReporter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if er := recover(); er != nil {
				err = grpc.Errorf(codes.Internal, "%v", er)
				r.ReportGRPCError(err, info, req)
			}
		}()
		return handler(ctx, req)
	}
}
