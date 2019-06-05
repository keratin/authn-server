package private

import (
	"net"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	keypackage "github.com/keratin/authn-server/app/data/private"
	authnpb "github.com/keratin/authn-server/grpc"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

func unsecuredServerSetup(t *testing.T, app *app.App) (authnpb.UnsecuredAdminAuthNClient, func()) {

	app.Config.UsernameIsEmail = true

	srvCtx := context.Background()
	srv := grpctest.NewPrivateServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}
	authnpb.RegisterUnsecuredAdminAuthNServer(srv, unsecuredServer{app})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	clientConn, err := grpc.Dial(conn.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewUnsecuredAdminAuthNClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}

func TestServiceConfiguration(t *testing.T) {

	app := &app.App{
		Config: &app.Config{
			AuthNURL: &url.URL{Scheme: "https", Host: "authn.example.com", Path: "/foo"},
		},
	}

	client, teardown := unsecuredServerSetup(t, app)
	defer teardown()

	res, err := client.ServiceConfiguration(context.TODO(), &authnpb.ServiceConfigurationRequest{})
	require.NoError(t, err)
	assert.Equal(t, "https://authn.example.com/foo/jwks", res.GetJwksUri())
}

func TestJWKS(t *testing.T) {
	rsaKey, err := keypackage.GenerateKey(512)
	require.NoError(t, err)

	testApp := test.App()
	testApp.KeyStore = mock.NewKeyStore(rsaKey)
	client, teardown := unsecuredServerSetup(t, testApp)
	defer teardown()

	res, err := client.JWKS(context.TODO(), &authnpb.JWKSRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}
