package public

import (
	"fmt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/internal/meta"
	context "golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type signupServiceServer struct {
	app *app.App
}

var _ authnpb.SignupServiceServer = signupServiceServer{}

func (s signupServiceServer) Signup(ctx context.Context, req *authnpb.SignupRequest) (*authnpb.SignupResponseEnvelope, error) {

	account, err := services.AccountCreator(s.app.AccountStore, s.app.Config, req.GetUsername(), req.GetPassword())
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	sessionToken, identityToken, err := services.SessionCreator(
		s.app.AccountStore, s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		account.ID, &s.app.Config.ApplicationDomains[0], meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	header := metadata.Pairs(s.app.Config.SessionCookieName, sessionToken)
	grpc.SendHeader(ctx, header)

	return &authnpb.SignupResponseEnvelope{
		Result: &authnpb.SignupResponse{
			IdToken: identityToken,
		},
	}, nil
}

func (s signupServiceServer) IsUsernameAvailable(ctx context.Context, req *authnpb.IsUsernameAvailableRequest) (*authnpb.IsUsernameAvailableResponseEnvelope, error) {
	account, err := s.app.AccountStore.FindByUsername(req.GetUsername())
	if err != nil {
		panic(err)
	}

	if account == nil {
		return &authnpb.IsUsernameAvailableResponseEnvelope{
			Result: true,
		}, nil
	}

	br := &errdetails.BadRequest{}
	br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
		Field:       "username",
		Description: services.ErrTaken,
	})

	statusError := status.New(codes.FailedPrecondition, services.FieldErrors{{Field: "username", Message: services.ErrTaken}}.Error())
	statusEr, e := statusError.WithDetails(br)
	if e != nil {
		panic(fmt.Sprintf("Unexpected error attaching metadata: %v", e))
	}
	return nil, statusEr.Err()
}
