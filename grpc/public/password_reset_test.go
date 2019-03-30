package public

import (
	"net"
	"testing"

	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func passwordResetServerSetup(t *testing.T, app *app.App) (authnpb.PasswordResetServiceClient, func()) {

	app.Config.UsernameIsEmail = true

	srvCtx := context.Background()
	srv := grpctest.NewPublicServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}
	authnpb.RegisterPasswordResetServiceServer(srv, passwordResetServer{app})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	clientConn, err := grpc.Dial(conn.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewPasswordResetServiceClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}
func TestRequestPasswordReset(t *testing.T) {

	testApp := test.App()
	client, teardown := passwordResetServerSetup(t, testApp)
	defer teardown()

	t.Run("known account", func(t *testing.T) {
		_, err := testApp.AccountStore.Create("known@keratin.tech", []byte("pwd"))
		require.NoError(t, err)

		req := &authnpb.PasswordResetRequest{
			Username: "known@keratin.tech",
		}
		_, err = client.RequestPasswordReset(context.TODO(), req)

		require.NoError(t, err)

		// TODO: assert go routine?
	})

	t.Run("unknown account", func(t *testing.T) {
		req := &authnpb.PasswordResetRequest{
			Username: "unknown@keratin.tech",
		}
		_, err := client.RequestPasswordReset(context.TODO(), req)

		require.NoError(t, err)
	})

}
