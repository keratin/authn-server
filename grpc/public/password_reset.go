package public

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	authnpb "github.com/keratin/authn-server/grpc"
	context "golang.org/x/net/context"
)

var _ authnpb.PasswordResetServiceServer = passwordResetServer{}

type passwordResetServer struct {
	app *app.App
}

func (s passwordResetServer) RequestPasswordReset(ctx context.Context, req *authnpb.PasswordResetRequest) (*authnpb.PasswordResetResponse, error) {

	account, err := s.app.AccountStore.FindByUsername(req.GetUsername())
	if err != nil {
		panic(err)
	}

	// run in the background so that a timing attack can't enumerate usernames
	go func() {
		err := services.PasswordResetSender(s.app.Config, account)
		if err != nil {
			s.app.Reporter.ReportError(err)
		}
	}()

	return &authnpb.PasswordResetResponse{}, nil
}
