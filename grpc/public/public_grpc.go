package public

import (
	"context"
	"net"

	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/meta"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func RunPublicGRPC(ctx context.Context, app *app.App, l net.Listener) error {
	opts := meta.PublicServerOptions(app)
	srv := grpc.NewServer(opts...)

	RegisterPublicGRPCMethods(srv, app)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	if err := srv.Serve(l); err != nil {
		log.Printf("serve error: %s", err)
		return err
	}
	return nil
}

func RegisterPublicGRPCMethods(srv *grpc.Server, app *app.App) {
	authnpb.RegisterPublicAuthNServer(srv, publicServer{
		app: app,
	})

	if app.Config.EnableSignup {
		authnpb.RegisterSignupServiceServer(srv, signupServiceServer{
			app: app,
		})
	}

	if app.Config.AppPasswordResetURL != nil {
		authnpb.RegisterPasswordResetServiceServer(srv, passwordResetServer{
			app: app,
		})
	}

	if app.Config.AppPasswordlessTokenURL != nil {
		authnpb.RegisterPasswordlessServiceServer(srv, passwordlessServer{
			app: app,
		})
	}
}
