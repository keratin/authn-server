package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/server/sessions"
	"google.golang.org/grpc/metadata"
)

type GatewayResponseMiddleware func(ctx context.Context, response http.ResponseWriter, m proto.Message) error

func JSONMarshaler() runtime.Marshaler {
	return new(customJSONMarshaler)
}

// StatusCodeMutator changes the HTTP Response status code to the desired mapping. The default mapping
// by gRPC-Gateway doesn't have some of the desired responses (e.g. 201), which is why this function is needed.
func StatusCodeMutator(ctx context.Context, response http.ResponseWriter, m proto.Message) error {
	switch m.(type) {
	case *authnpb.SignupResponseEnvelope, *authnpb.LoginResponseEnvelope, *authnpb.RefreshSessionResponseEnvelope, *authnpb.SubmitPasswordlessLoginResponseEnvelope, *authnpb.ChangePasswordResponseEnvelope:
		response.WriteHeader(http.StatusCreated)
	}
	return nil
}

// CookieSetter extracts the session cookie from metadata and assigns it to a cookie. If the session
// value is an empty string, then the cookie is marked to be removed.
func CookieSetter(cfg *app.Config) GatewayResponseMiddleware {
	return func(ctx context.Context, response http.ResponseWriter, m proto.Message) error {
		switch m.(type) {
		case *authnpb.LogoutResponse, *authnpb.SignupResponseEnvelope, *authnpb.LoginResponseEnvelope:
			md, ok := runtime.ServerMetadataFromContext(ctx)
			if !ok {
				return fmt.Errorf("Failed to extract ServerMetadata from context")
			}
			ss := md.HeaderMD.Get(cfg.SessionCookieName)
			if len(ss) != 1 {
				return fmt.Errorf("Received more than a single session value")
			}
			sessions.Set(cfg, response, ss[0])
		}
		return nil
	}
}

// CookieAnnotator reads the cookie from the *http.Request and sends it to the gRPC service as gRPC metadata
func CookieAnnotator(app *app.App) func(ctx context.Context, req *http.Request) metadata.MD {
	return func(ctx context.Context, req *http.Request) metadata.MD {
		cookie, err := req.Cookie(app.Config.SessionCookieName)
		if err != nil {
			app.Reporter.ReportRequestError(err, req)
			return nil
		}
		return metadata.MD{
			app.Config.SessionCookieName: []string{cookie.Value},
		}
	}
}

// FormWrapper takes form values from application/x-www-form-urlencoded and converts it to JSON
// Workaround adapted from: https://github.com/grpc-ecosystem/grpc-gateway/issues/7#issuecomment-358569373
func FormWrapper(mux http.Handler) http.Handler {
	// We can't use strconv#ParseBool because
	// our assumption for `true` values don't match.
	// strconv.ParseBool considers the following as `true`: 1, t, T, TRUE, true, True
	boolRegexp := regexp.MustCompile("^(?i:t|true|yes)$")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(strings.Split(r.Header.Get("Content-Type"), ";")[0]) == "application/x-www-form-urlencoded" {
			if err := r.ParseForm(); err != nil {
				// deliberately omit including the form from
				// logging to avoid leaking sensetive data
				log.WithError(err).WithFields(log.Fields{
					"url":    r.URL,
					"method": r.Method,
				}).Warn("error parsing form")
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			jsonMap := make(map[string]interface{}, len(r.Form))
			for k, v := range r.Form {
				if len(v) > 0 {
					/*
					* We have a special case of converting form data to JSON. Our
					* type conversion is only concerned with limited set of keys,
					* so it's simpler to just check for these and convert the values
					* if those exist.
					**/
					switch k {
					case "id":
						val, err := strconv.ParseInt(v[0], 10, 64)
						if err != nil {
							log.WithError(err).WithFields(log.Fields{
								"url":    r.URL,
								"method": r.Method,
							}).Warn("error converting field `id` to int64")
							http.Error(w, "", http.StatusBadRequest)
							return
						}
						jsonMap[k] = val
						continue

					case "locked":
						jsonMap[k] = boolRegexp.MatchString(v[0])
						continue

					default:
						jsonMap[k] = v[0]
					}
				}
			}
			jsonBody, err := json.Marshal(jsonMap)
			if err != nil {
				// deliberately omit including the form from
				// logging to avoid leaking sensetive data
				log.WithError(err).WithFields(log.Fields{
					"url":    r.URL,
					"method": r.Method,
				}).Warn("error encoding form into JSON")
				http.Error(w, "", http.StatusBadRequest)
				return
			}
			r.Body = ioutil.NopCloser(bytes.NewReader(jsonBody))
			r.ContentLength = int64(len(jsonBody))
			r.Header.Set("Content-Type", "application/json")
		}
		mux.ServeHTTP(w, r)
	})
}
