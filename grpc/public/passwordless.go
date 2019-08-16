package public

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/internal/meta"
	context "golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

type passwordlessServer struct {
	app *app.App
}

var _ authnpb.PasswordlessServiceServer = passwordlessServer{}

func (s passwordlessServer) RequestPasswordlessLogin(ctx context.Context, req *authnpb.RequestPasswordlessLoginRequest) (*authnpb.RequestPasswordlessLoginResponse, error) {

	account, err := s.app.AccountStore.FindByUsername(req.GetUsername())
	if err != nil {
		panic(err)
	}

	// run in the background so that a timing attack can't enumerate usernames
	go func() {
		err := services.PasswordlessTokenSender(s.app.Config, account)
		if err != nil {
			info := meta.GetUnaryServerInfo(ctx)
			s.app.Reporter.ReportGRPCError(err, info, req)
		}
	}()

	return &authnpb.RequestPasswordlessLoginResponse{}, nil
}

func (s passwordlessServer) SubmitPasswordlessLogin(ctx context.Context, req *authnpb.SubmitPasswordlessLoginRequest) (*authnpb.SubmitPasswordlessLoginResponseEnvelope, error) {

	var err error
	var accountID int

	accountID, err = services.PasswordlessTokenVerifier(
		s.app.AccountStore,
		s.app.Reporter,
		s.app.Config,
		req.GetToken(),
	)

	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	sessionToken, identityToken, err := services.SessionCreator(
		s.app.AccountStore, s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		accountID, meta.MatchedDomain(ctx), meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	meta.SetSession(ctx, s.app.Config.SessionCookieName, sessionToken)

	// Return the signed identity token in the body
	return &authnpb.SubmitPasswordlessLoginResponseEnvelope{
		Result: &authnpb.SubmitPasswordlessLoginResponse{
			IdToken: identityToken,
		},
	}, nil
}
