package public

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/gateway"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func RegisterPublicGatewayHandlers(ctx context.Context, app *app.App, r *mux.Router, mux *runtime.ServeMux, conn *grpc.ClientConn) {
	authnpb.RegisterPublicAuthNHandler(ctx, mux, conn)
	if app.Config.EnableSignup {
		authnpb.RegisterSignupServiceHandler(ctx, mux, conn)
	}

	if app.Config.AppPasswordResetURL != nil {
		authnpb.RegisterPasswordResetServiceHandler(ctx, mux, conn)
	}

	if app.Config.AppPasswordlessTokenURL != nil {
		authnpb.RegisterPasswordlessServiceHandler(ctx, mux, conn)
	}
}

func RunPublicGateway(ctx context.Context, app *app.App, r *mux.Router, conn *grpc.ClientConn, l net.Listener) error {

	gmux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(gateway.CookieSetter(app.Config)), // Cookies always have to go first
		runtime.WithForwardResponseOption(gateway.StatusCodeMutator),
		runtime.WithMarshalerOption("*", &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
	)

	RegisterRoutes(r, app, gmux)
	RegisterPublicGatewayHandlers(ctx, app, r, gmux, conn)
	s := &http.Server{
		Addr:    l.Addr().String(),
		Handler: gateway.WrapRouter(gateway.FormWrapper(r), app),
	}

	go func() {
		<-ctx.Done()
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http server: %v", err)
		}
	}()

	if err := s.Serve(l); err != http.ErrServerClosed {
		return err
	}

	return nil
}
