package grpc

import (
	context "context"
	fmt "fmt"
	"net/http"
	"net/textproto"

	proto "github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/grpclog"
)

func init() {
	// public
	forward_SignupService_Signup_0 = enveloper
	forward_PasswordlessService_SubmitPasswordlessLogin_0 = enveloper
	forward_PublicAuthN_Login_0 = enveloper
	forward_PublicAuthN_RefreshSession_0 = enveloper
	forward_PublicAuthN_ChangePassword_0 = enveloper

	// private
	forward_SecuredAdminAuthN_GetAccount_0 = enveloper
	forward_SecuredAdminAuthN_ImportAccount_0 = enveloper
}

// The code from this line onwards is based on the code found in the gRPC-Gateway project with minor modifications.
// Base: https://github.com/grpc-ecosystem/grpc-gateway/blob/c677e419aa5ce4a305e17248df3c91b6a3955081/runtime/handler.go#L116

// enveloper wraps the gRPC response with an envelope of the form:
// {
//		"result": <message>
// }
func enveloper(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
	}

	handleForwardResponseServerMetadata(w, mux, md)
	handleForwardResponseTrailerHeader(w, md)

	contentType := marshaler.ContentType()
	// Check marshaler on run time in order to keep backwards compatability
	// An interface param needs to be added to the ContentType() function on
	// the Marshal interface to be able to remove this check
	if httpBodyMarshaler, ok := marshaler.(*runtime.HTTPBodyMarshaler); ok {
		contentType = httpBodyMarshaler.ContentTypeFromMessage(resp)
	}
	w.Header().Set("Content-Type", contentType)

	if err := handleForwardResponseOptions(ctx, w, resp, opts); err != nil {
		runtime.HTTPError(ctx, mux, marshaler, w, req, err)
		return
	}
	var buf []byte
	var err error
	buf, err = marshaler.Marshal(map[string]interface{}{
		"result": resp,
	})

	if err != nil {
		grpclog.Infof("Marshal error: %v", err)
		runtime.HTTPError(ctx, mux, marshaler, w, req, err)
		return
	}

	if _, err = w.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}

	handleForwardResponseTrailer(w, md)
}

func handleForwardResponseServerMetadata(w http.ResponseWriter, mux *runtime.ServeMux, md runtime.ServerMetadata) {
	for k, vs := range md.HeaderMD {
		if h, ok := runtime.DefaultHeaderMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}

func handleForwardResponseOptions(ctx context.Context, w http.ResponseWriter, resp proto.Message, opts []func(context.Context, http.ResponseWriter, proto.Message) error) error {
	if len(opts) == 0 {
		return nil
	}
	for _, opt := range opts {
		if err := opt(ctx, w, resp); err != nil {
			grpclog.Infof("Error handling ForwardResponseOptions: %v", err)
			return err
		}
	}
	return nil
}
