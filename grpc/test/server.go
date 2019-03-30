package test

import (
	"context"
	"net"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/grpc/internal/meta"
	"google.golang.org/grpc"
)

func NewPublicServer(app *app.App) *grpc.Server {
	opts := meta.PublicServerOptions(app)
	return grpc.NewServer(opts...)
}

func NewPrivateServer(app *app.App) *grpc.Server {
	opts := meta.PrivateServerOptions(app)
	return grpc.NewServer(opts...)
}

func Serve(ctx context.Context, srv *grpc.Server, l net.Listener) {
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	srv.Serve(l)
}
