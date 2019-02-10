package private

import (
	"encoding/json"

	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/lib/compat"
	"golang.org/x/net/context"
	jose "gopkg.in/square/go-jose.v2"
)

type unsecuredServer struct {
	app *app.App
}

func (ss unsecuredServer) ServiceConfiguration(context.Context, *authnpb.ServiceConfigurationRequest) (*authnpb.Configuration, error) {
	return &authnpb.Configuration{
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
		keyID, err := compat.KeyID(key.Public())
		if err != nil {
			ss.app.Reporter.ReportError(err)
		} else {
			// There are no proto definitions for jose.JSONWebKey and the marshalled version
			// looks different than the struct, so the workaround is to build jose.JSONWebKey,
			// marshal it , then unmarshal it into our message.
			k, err := jose.JSONWebKey{
				Key:       key.Public(),
				Use:       "sig",
				Algorithm: "RS256",
				KeyID:     keyID,
			}.MarshalJSON()
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
	}
	return &authnpb.JWKSResponse{
		Keys: keys,
	}, nil
}
