package public

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/internal/meta"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type signupServiceServer struct {
	app *app.App
}

var _ authnpb.SignupServiceServer = signupServiceServer{}

func (s signupServiceServer) Signup(ctx context.Context, req *authnpb.SignupRequest) (*authnpb.SignupResponse, error) {

	account, err := services.AccountCreator(s.app.AccountStore, s.app.Config, req.GetUsername(), req.GetPassword())
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	sessionToken, identityToken, err := services.SessionCreator(
		s.app.AccountStore, s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		account.ID, meta.MatchedDomain(ctx), meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	header := metadata.Pairs(s.app.Config.SessionCookieName, sessionToken)
	grpc.SendHeader(ctx, header)

	return &authnpb.SignupResponse{
		IdToken: identityToken,
	}, nil
}

func (s signupServiceServer) IsUsernameAvailable(ctx context.Context, req *authnpb.IsUsernameAvailableRequest) (*authnpb.IsUsernameAvailableResponse, error) {
	account, err := s.app.AccountStore.FindByUsername(req.GetUsername())
	if err != nil {
		panic(err)
	}

	if account == nil {
		return &authnpb.IsUsernameAvailableResponse{
			Result: true,
		}, nil
	}
	return nil, errors.ToStatusErrorWithDetails(services.FieldErrors{{Field: "username", Message: services.ErrTaken}}, codes.FailedPrecondition).Err()
}
