package private

import (
	"encoding/base64"
	"net"
	"testing"

	"google.golang.org/grpc/credentials"

	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/server/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

type basicAuth struct {
	username string
	password string
}

func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (basicAuth) RequireTransportSecurity() bool {
	return true
}

func statServerSetup(t *testing.T, app *app.App) (authnpb.AuthNActivesClient, func()) {

	app.Config.UsernameIsEmail = true

	srvCtx := context.Background()
	srv := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(&grpctest.Cert)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.StandardLogger())),
			grpc_prometheus.UnaryServerInterceptor,
			// the default authentication is none
			grpc_auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
				return ctx, nil
			}),
		),
	) // grpctest.NewPrivateServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}

	authnpb.RegisterAuthNActivesServer(srv, statsServer{app: app, matcher: func(username, password string) bool { return true }})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	// error handling omitted

	clientConn, err := grpc.Dial(
		conn.Addr().String(),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(grpctest.CertPool, "")),
		grpc.WithPerRPCCredentials(basicAuth{
			username: app.Config.AuthUsername,
			password: app.Config.AuthPassword,
		}),
	)
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewAuthNActivesClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}

func TestServiceStats(t *testing.T) {
	app := test.App()
	client, teardown := statServerSetup(t, app)
	defer teardown()

	err := app.Actives.Track(1)
	require.NoError(t, err)

	res, err := client.ServiceStats(context.TODO(), &authnpb.ServiceStatsRequest{})
	require.NoError(t, err)
	t.Log(res)
	assert.NotEmpty(t, res)
}
