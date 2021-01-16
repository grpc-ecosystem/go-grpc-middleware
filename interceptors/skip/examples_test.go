package skip_test

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/skip"
)

// Simple example of skipping auth interceptor in the reflection method.
func Example_initialization() {
	_ = grpc.NewServer(
		grpc.UnaryInterceptor(skip.UnaryServerInterceptor(auth.UnaryServerInterceptor(dummyAuth), SkipReflectionService)),
		grpc.StreamInterceptor(skip.StreamServerInterceptor(auth.StreamServerInterceptor(dummyAuth), SkipReflectionService)),
	)
}

func dummyAuth(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func SkipReflectionService(_ context.Context, _ interceptors.GRPCType, service string, _ string) bool {
	return service != "grpc.reflection.v1alpha.ServerReflection"
}
