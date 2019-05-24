package meta

import (
	"context"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/ops"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

// PrivateServerOptions provides the default set of gRPC server options for the private interface
func PrivateServerOptions(app *app.App) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			ops.GRPCRecoveryInterceptor(app.Reporter),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.StandardLogger())),
			grpc_prometheus.UnaryServerInterceptor,
			// the default authentication is none
			grpc_auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
				return ctx, nil
			}),
			SessionInterceptor(app),
		),
	}
}

// PublicServerOptions provides the default set of gRPC server options for the public interface
func PublicServerOptions(app *app.App) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			ops.GRPCRecoveryInterceptor(app.Reporter),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.StandardLogger())),
			grpc_prometheus.UnaryServerInterceptor,
			SessionInterceptor(app),
		),
	}
}
