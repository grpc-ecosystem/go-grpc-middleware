// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package auth_test

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

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

// Simple example of server initialization code
func Example_serverConfig() {
	_ = grpc.NewServer(
		grpc.StreamInterceptor(auth.StreamServerInterceptor(exampleAuthFunc)),
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(exampleAuthFunc)),
	)
}

type gRPCServerAuthenticated struct {
	pb.UnimplementedGreeterServer
}

// SayHello only can be called by client when authenticated by exampleAuthFunc
func (g gRPCServerAuthenticated) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "pong authenticated"}, nil
}

type gRPCServerUnauthenticated struct {
	pb.UnimplementedGreeterServer
}

// SayHello can be called by client without being authenticated by exampleAuthFunc as AuthFuncOverride is called instead
func (g *gRPCServerUnauthenticated) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "pong unauthenticated"}, nil
}

// AuthFuncOverride is called instead of exampleAuthFunc
func (g *gRPCServerUnauthenticated) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	log.Println("client is calling method:", fullMethodName)
	return ctx, nil
}

// Simple example of server initialization code with AuthFuncOverride method.
func Example_serverConfigWithAuthOverride() {
	server := grpc.NewServer(
		grpc.StreamInterceptor(auth.StreamServerInterceptor(exampleAuthFunc)),
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(exampleAuthFunc)),
	)

	overrideActive := true

	if overrideActive {
		pb.RegisterGreeterServer(server, &gRPCServerUnauthenticated{})
	} else {
		pb.RegisterGreeterServer(server, &gRPCServerAuthenticated{})
	}
}
