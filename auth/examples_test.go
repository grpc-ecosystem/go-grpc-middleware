package grpc_auth_test

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/status"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

func parseToken(token string) (struct{}, error) {
	return struct{}{}, nil
}

func userClaimFromToken(struct{}) string {
	return "foobar"
}

// exampleAuthFunc is used by a middleware to authenticate requests
func exampleAuthFunc(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	tokenInfo, err := parseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}

	grpc_ctxtags.Extract(ctx).Set("auth.sub", userClaimFromToken(tokenInfo))

	// WARNING: in production define your own type to avoid context collisions
	newCtx := context.WithValue(ctx, "tokenInfo", tokenInfo)

	return newCtx, nil
}

// Simple example of server initialization code
func Example_serverConfig() {
	_ = grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(exampleAuthFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(exampleAuthFunc)),
	)
}

type gRPCserverAuthenticated struct{}

// SayHello only can be called by client when authenticated by exampleAuthFunc
func (g gRPCserverAuthenticated) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "pong authenticated"}, nil
}

type gRPCserverUnauthenticated struct{}

// SayHello can be called by client without being authenticated by exampleAuthFunc as AuthFuncOverride is called instead
func (g *gRPCserverUnauthenticated) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "pong unauthenticated"}, nil
}

// AuthFuncOverride is called instead of exampleAuthFunc
func (g *gRPCserverUnauthenticated) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	log.Println("client is calling method:", fullMethodName)
	return ctx, nil
}

// Simple example of server initialization code with AuthFuncOverride method.
func Example_serverConfigWithAuthOverride() {
	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(exampleAuthFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(exampleAuthFunc)),
	)

	overrideActive := true

	if overrideActive {
		pb.RegisterGreeterServer(server, &gRPCserverUnauthenticated{})
	} else {
		pb.RegisterGreeterServer(server, &gRPCserverAuthenticated{})
	}
}
