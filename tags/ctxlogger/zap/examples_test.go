package ctxlogger_zap_test

import (
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"

	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/ctxlogger/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization(zapLogger *zap.Logger) *grpc.Server {
	// Create a server, make sure we put the grpc_ctxtags context before everything else.
	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			ctxlogger_zap.UnaryServerInterceptor(zapLogger),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			ctxlogger_zap.StreamServerInterceptor(zapLogger),
		),
	)
	return server
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func Example_handlerUsageUnaryPing() interface{} {
	x := func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

		// Extract a single request-scoped zap.Logger and log messages.
		l := ctxlogger_zap.Extract(ctx)
		l.Info("some ping")
		l.Info("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
	return x
}
