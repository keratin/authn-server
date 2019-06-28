package errors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/server/handlers"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const wwwAuthenticate = `WWW-Authenticate`

func ToFieldErrors(errs *errdetails.BadRequest) services.FieldErrors {
	fes := services.FieldErrors{}
	for _, violation := range errs.GetFieldViolations() {
		fes = append(fes, services.FieldError{Field: violation.GetField(), Message: violation.GetDescription()})
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
	authen := md.HeaderMD.Get(wwwAuthenticate)

	statusError := status.Convert(err)

	// Basic-Auth failure
	if statusError.Code() == codes.Unauthenticated && len(authen) > 0 {
		w.Header().Set(wwwAuthenticate, `Basic realm="Private AuthN Realm"`)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized.\n"))
		return
	}

	// Non-Basic-Auth authentication failure (only possible in /session/refresh)
	if statusError.Code() == codes.Unauthenticated {
		w.WriteHeader(401)
		return
	}

	for _, detail := range statusError.Details() {

		switch t := detail.(type) {
		case *errdetails.BadRequest:
			// Convert the errors back to AuthN's custom error responses to preserve the shape of the returned error
			fes := ToFieldErrors(t)
			if statusError.Code() == codes.NotFound {
				handlers.WriteNotFound(w, fes[0].Field)
			} else {
				handlers.WriteErrors(w, fes)
			}

		default:
			// Fallback to the default error handler
			runtime.DefaultHTTPError(ctx, mux, marshaler, w, req, err)
		}
	}
}
