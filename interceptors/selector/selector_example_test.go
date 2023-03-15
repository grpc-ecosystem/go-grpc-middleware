// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package selector_test

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
)

// alwaysPassLimiter is an example limiter which implements Limiter interface.
// It does not limit any request because Limit function always returns false.
type alwaysPassLimiter struct{}

func (*alwaysPassLimiter) Limit(_ context.Context) error {
	return nil
}

func healthSkip(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != "/ping.v1.PingService/Health"
}

func Example_ratelimit() {
	limiter := &alwaysPassLimiter{}
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			selector.UnaryServerInterceptor(ratelimit.UnaryServerInterceptor(limiter), selector.MatchFunc(healthSkip)),
		),
		grpc.ChainStreamInterceptor(
			selector.StreamServerInterceptor(ratelimit.StreamServerInterceptor(limiter), selector.MatchFunc(healthSkip)),
		),
	)
}

var tokenInfoKey struct{}

func parseToken(token string) (struct{}, error) {
	return struct{}{}, nil
}

func userClaimFromToken(struct{}) string {
	return "foobar"
}

// exampleAuthFunc is used by a middleware to authenticate requests
func exampleAuthFunc(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	tokenInfo, err := parseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", userClaimFromToken(tokenInfo)})

	// WARNING: In production define your own type to avoid context collisions.
	return context.WithValue(ctx, tokenInfoKey, tokenInfo), nil
}

func loginSkip(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != "/auth.v1.AuthService/Login"
}

func Example_login() {
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(exampleAuthFunc), selector.MatchFunc(loginSkip)),
		),
		grpc.ChainStreamInterceptor(
			selector.StreamServerInterceptor(auth.StreamServerInterceptor(exampleAuthFunc), selector.MatchFunc(loginSkip)),
		),
	)
}
