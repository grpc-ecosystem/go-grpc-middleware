package grpc_jwt_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/golang-jwt/jwt/v4"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_jwt "github.com/grpc-ecosystem/go-grpc-middleware/auth/jwt"
	pb "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
)

var (
	exampleClaims = jwt.MapClaims{
		"foo": "bar",
		"nbf": float64(time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix()),
	}
)

// Simple example of server initialization code
func Example_server() {
	jwtAuthFunc := grpc_jwt.NewAuthFunc([]byte("some_secret"))
	svr := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(jwtAuthFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(jwtAuthFunc)),
	)
	service := &service{}
	pb.RegisterTestServiceServer(svr, service)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, exampleClaims)
	signedToken, _ := token.SignedString([]byte("some_secret"))
	service.Ping(ctxWithToken(context.TODO(), "bearer", signedToken), &pb.PingRequest{Value: "some_value"})
}

// Example of server initialization code
func Example_serverConfig() {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	jwtAuthFunc := grpc_jwt.NewAuthFuncWithConfig(
		grpc_jwt.Config{
			SigningMethod: jwt.SigningMethodES256.Name,
			SigningKey:    key.PublicKey,
		},
	)
	svr := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(jwtAuthFunc)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(jwtAuthFunc)),
	)
	service := &service{}
	pb.RegisterTestServiceServer(svr, service)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, exampleClaims)
	signedToken, _ := token.SignedString(key)
	service.Ping(ctxWithToken(context.TODO(), "bearer", signedToken), nil)
}

type service struct {
	pb.UnimplementedTestServiceServer
}

// SayHello can only be called by client when authenticated by jwtAuthFunc
func (g *service) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingResponse, error) {
	token := ctx.Value(grpc_jwt.DefaultContextKey).(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	return &pb.PingResponse{Value: fmt.Sprintf("pong with claim foo='%v': %v", claims["foo"], request.Value)}, nil
}
