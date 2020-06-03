package zerolog_test

import (
	"context"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	grpc_zerolog "github.com/irridia/go-grpc-middleware/logging/zerolog"
	"github.com/irridia/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	zerologLogger zerolog.Logger = zerolog.New(os.Stderr)
	customFunc    grpc_zerolog.CodeToLevel
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization() {
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithLevels(customFunc),
	}
	// Make sure that log statements internal to gRPC library are logged using the zerolog Logger as well.
	grpc_zerolog.ReplaceGrpcLoggerV2(&zerologLogger)
	// Create a server, make sure we put the grpc_ctxtags context before everything else.
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zerolog.UnaryServerInterceptor(&zerologLogger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zerolog.StreamServerInterceptor(&zerologLogger, opts...),
		),
	)
}

func Example_initializationWithDurationFieldOverride() {
	// Shared options for the logger, with a custom duration to log field function.
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(&zerologLogger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(&zerologLogger, opts...),
		),
	)
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
		// Extract a single request-scoped zerolog.Logger and log messages.
		l := ctxzerolog.Extract(ctx).Logger()
		l.Info().Msg("some ping")
		l.Info().Msg("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}

func ExampleWithDecider() {
	opts := []grpc_zerolog.Option{
		grpc_zerolog.WithDecider(func(methodFullName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && methodFullName == "blah.foo.healthcheck" {
				return false
			}

			// by default you will log all calls
			return true
		}),
	}

	_ = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zerolog.StreamServerInterceptor(&zerologLogger, opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zerolog.UnaryServerInterceptor(&zerologLogger, opts...)),
	}
}
