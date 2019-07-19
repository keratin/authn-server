package private

import (
	"crypto/subtle"
	"encoding/base64"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/meta"
	"github.com/keratin/authn-server/grpc/public"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type basicAuthMatcher func(username, password string) bool

// RunPrivateGRPC registers the private services and runs the gRPC server on the provided listener
func RunPrivateGRPC(ctx context.Context, app *app.App, l net.Listener) error {
	opts := meta.PrivateServerOptions(app)
	srv := grpc.NewServer(opts...)

	public.RegisterPublicGRPCMethods(srv, app)

	matcher := func(u string, p string) bool {
		usernameMatch := subtle.ConstantTimeCompare([]byte(u), []byte(app.Config.AuthUsername))
		passwordMatch := subtle.ConstantTimeCompare([]byte(p), []byte(app.Config.AuthPassword))

		return usernameMatch == 1 && passwordMatch == 1
	}

	authnpb.RegisterSecuredAdminAuthNServer(srv, securedServer{
		app:     app,
		matcher: matcher,
	})

	authnpb.RegisterUnsecuredAdminAuthNServer(srv, unsecuredServer{
		app: app,
	})

	if app.Actives != nil {
		authnpb.RegisterAuthNActivesServer(srv, statsServer{
			app:     app,
			matcher: matcher,
		})
	}

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

func basicAuthCheck(ctx context.Context, matcher basicAuthMatcher) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "basic")
	if err != nil {
		header := metadata.Pairs("WWW-authenticate", `Basic realm="Private AuthN Realm"`)
		grpc.SendHeader(ctx, header)
		return ctx, grpc.Errorf(codes.Unauthenticated, "missing context metadata")
	}

	c, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		header := metadata.Pairs("WWW-authenticate", `Basic realm="Private AuthN Realm"`)
		grpc.SendHeader(ctx, header)
		return ctx, status.Error(codes.Unauthenticated, `invalid base64 in header`)
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		header := metadata.Pairs("WWW-authenticate", `Basic realm="Private AuthN Realm"`)
		grpc.SendHeader(ctx, header)
		return ctx, status.Error(codes.Unauthenticated, `invalid basic auth format`)
	}

	user, password := cs[:s], cs[s+1:]
	if !matcher(user, password) {
		err := grpc.SetHeader(ctx, metadata.Pairs("www-authenticate", `Basic realm="Private AuthN Realm"`))
		if err != nil {
			log.Errorf("error setting header: %s", err)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "WWW-Authenticate", `Basic realm="Private AuthN Realm"`)
		return ctx, status.Error(codes.Unauthenticated, "invalid user or password")
	}
	return ctx, nil
}
