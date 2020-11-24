package skip_test

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/skip"
)

// Simple example of skipping auth interceptor in the reflection method.
func Example_initialization() {
	_ = grpc.NewServer(
		grpc.UnaryInterceptor(skip.UnaryServerInterceptor(auth.UnaryServerInterceptor(exampleAuthFunc), ReflectionFilter)),
		grpc.StreamInterceptor(skip.StreamServerInterceptor(auth.StreamServerInterceptor(exampleAuthFunc), ReflectionFilter)),
	)
}

func exampleAuthFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func ReflectionFilter(ctx context.Context, gRPCType interceptors.GRPCType, service string, method string) bool {
	return service == "grpc.reflection.v1alpha.ServerReflection"
}
