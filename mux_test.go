package grpc_middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	methodA1 = "serviceA.Method1"
	methodA2 = "serviceA.Method2"
	methodB1 = "serviceB.Method1"
)

func TestMuxUnaryServer(t *testing.T) {
	input := "input"

	first := pathInterceptor("first")
	second := pathInterceptor("second")
	third := pathInterceptor("third")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return ctx.Value("path"), nil
	}

	mux := MuxUnaryServer(UnaryMux{methodA1: first, methodA2: second, "default": third})

	tests := [][]string{
		{methodA1, "parent/first"},
		{methodA2, "parent/second"},
		{methodB1, "parent/third"},
	}
	for _, tt := range tests {
		out, _ := mux(context.WithValue(parentContext, "path", "parent"), input, &grpc.UnaryServerInfo{FullMethod: tt[0]}, handler)
		require.EqualValues(t, tt[1], out, "mux path is incorrect")
	}
}

func pathInterceptor(node string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		path := ctx.Value("path")
		if str, ok := path.(string); ok {
			ctx = context.WithValue(ctx, "path", str+"/"+node)
		} else {
			ctx = context.WithValue(ctx, "path", "previous path not a string at "+node)
		}
		return handler(ctx, req)
	}
}
