package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app/services"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const wwwAuthneticate = `WWW-Authenticate`

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e FieldError) String() string {
	return fmt.Sprintf("%v: %v", e.Field, e.Message)
}

type FieldErrors []FieldError

type ServiceErrors struct {
	Errors FieldErrors `json:"errors"`
}

func (es FieldErrors) Error() string {
	var buf = make([]string, len(es))
	for i, e := range es {
		buf[i] = e.String()
	}
	return strings.Join(buf, ", ")
}

func ToFieldErrors(errs *errdetails.BadRequest) FieldErrors {
	fes := FieldErrors{}
	for _, violation := range errs.GetFieldViolations() {
		fes = append(fes, FieldError{Field: violation.GetField(), Message: violation.GetDescription()})
	}
	return fes
}

func ToStatusErrorWithDetails(fes services.FieldErrors, errCode codes.Code) *status.Status {
	br := &errdetails.BadRequest{}
	for _, fe := range fes {
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       fe.Field,
			Description: fe.Message,
		})
	}
	statusError := status.New(errCode, fes.Error())
	statusEr, e := statusError.WithDetails(br)
	if e != nil {
		panic(fmt.Sprintf("Unexpected error attaching details to error: %v", e))
	}
	return statusEr
}

// CustomHTTPError is a custom error handler to write the error JSON obeying the pre-defined error JSON structure
func CustomHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {

	md, _ := runtime.ServerMetadataFromContext(ctx)
	authen := md.HeaderMD.Get(wwwAuthneticate)

	statusError := status.Convert(err)

	// Basic-Auth failure
	if statusError.Code() == codes.Unauthenticated && len(authen) > 0 {
		w.Header().Set(wwwAuthneticate, `Basic realm="Private AuthN Realm"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized.\n"))
		return
	}

	for _, detail := range statusError.Details() {

		switch t := detail.(type) {
		case *errdetails.BadRequest:

			if statusError.Code() == codes.NotFound {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
			}

			// Convert the errors back to AuthN's custom error responses to preserve the shape of the returned error
			fes := ToFieldErrors(t)
			j, er := json.Marshal(ServiceErrors{fes})
			if er != nil {
				panic(er)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(j)

		default:
			// Fallback to the default error handler
			runtime.DefaultHTTPError(ctx, mux, marshaler, w, req, err)
		}
	}
}
