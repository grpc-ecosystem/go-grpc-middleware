// gRPC Server Interceptor routing middleware.

package grpc_middleware

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type UnaryMux map[string]grpc.UnaryServerInterceptor

// MuxUnaryServer creates a single interceptor out of a map of RPC to interceptors.
func MuxUnaryServer(mux UnaryMux) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if i, ok := mux[info.FullMethod]; ok {
			return i(ctx, req, info, handler)
		}
		if i, ok := mux["default"]; ok {
			return i(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}
