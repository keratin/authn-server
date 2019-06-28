package private

import (
	"encoding/json"

	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"golang.org/x/net/context"
)

type unsecuredServer struct {
	app *app.App
}

func (ss unsecuredServer) ServiceConfiguration(context.Context, *authnpb.ServiceConfigurationRequest) (*authnpb.ServiceConfigurationResponse, error) {
	return &authnpb.ServiceConfigurationResponse{
		Issuer:                           ss.app.Config.AuthNURL.String(),
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IdTokenSigningAlgValuesSupported: []string{"RS256"},
		ClaimsSupported:                  []string{"iss", "sub", "aud", "exp", "iat", "auth_time"},
		JwksUri:                          ss.app.Config.AuthNURL.String() + "/jwks",
	}, nil
}

func (ss unsecuredServer) JWKS(ctx context.Context, _ *authnpb.JWKSRequest) (*authnpb.JWKSResponse, error) {
	keys := []*authnpb.Key{}
	for _, key := range ss.app.KeyStore.Keys() {
		// There are no proto definitions for jose.JSONWebKey and the marshalled version
		// looks different than the struct, so the workaround is to build jose.JSONWebKey,
		// marshal it , then unmarshal it into our message.
		k, err := key.JWK.MarshalJSON()
		if err != nil {
			ss.app.Reporter.ReportError(err)
			continue
		}
		pkey := &authnpb.Key{}
		err = json.Unmarshal(k, pkey)
		if err != nil {
			ss.app.Reporter.ReportError(err)
			continue
		}
		keys = append(keys, pkey)
	}
	return &authnpb.JWKSResponse{
		Keys: keys,
	}, nil
}
