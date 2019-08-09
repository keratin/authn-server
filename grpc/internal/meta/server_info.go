package meta

import (
	"context"

	"google.golang.org/grpc"
)

type unaryServerInfo int

func unaryServerInfoInjector(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = context.WithValue(ctx, unaryServerInfo(0), info)
	return handler(ctx, req)
}

// GetUnaryServerInfo extracts the *grpc.UnaryServerInfo from the context.Context
func GetUnaryServerInfo(ctx context.Context) *grpc.UnaryServerInfo {
	val := ctx.Value(unaryServerInfo(0)).(*grpc.UnaryServerInfo)
	return val
}
