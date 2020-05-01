package auth_test

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

func parseToken(string) (struct{}, error) {
	return struct{}{}, nil
}

func userClaimFromToken(struct{}) string {
	return "foobar"
}

// Simple example of server initialization code.
func Example_serverConfig() {
	exampleAuthFunc := func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		tokenInfo, err := parseToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}
		tags.Extract(ctx).Set("auth.sub", userClaimFromToken(tokenInfo))
		newCtx := context.WithValue(ctx, "tokenInfo", tokenInfo)
		return newCtx, nil
	}

	_ = grpc.NewServer(
		grpc.StreamInterceptor(auth.StreamServerInterceptor(exampleAuthFunc)),
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(exampleAuthFunc)),
	)
}
