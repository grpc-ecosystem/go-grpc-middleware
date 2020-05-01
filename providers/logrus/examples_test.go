package logrus_test

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	grpclogrus "github.com/grpc-ecosystem/go-grpc-middleware/v2/providers/logrus"
)

var (
	customFunc             logging.CodeToLevel
	customDurationToFields logging.DurationToFields
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := logrus.New()
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []logging.Option{
		logging.WithLevels(customFunc),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.UnaryServerInterceptor(grpclogrus.InterceptorLogger(logger), opts...),
		),
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(tags.WithFieldExtractor(tags.CodeGenRequestFieldExtractor)),
			logging.StreamServerInterceptor(grpclogrus.InterceptorLogger(logger), opts...),
		),
	)
}

func Example_initializationWithDurationFieldOverride() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := logrus.New()
	// Shared options for the logger, with a custom duration to log field function.
	opts := []logging.Option{
		logging.WithDurationField(customDurationToFields),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(grpclogrus.InterceptorLogger(logger), opts...),
		),
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(),
			logging.StreamServerInterceptor(grpclogrus.InterceptorLogger(logger), opts...),
		),
	)
}

func ExampleWithDecider() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := logrus.New()
	// Shared options for the logger, with a custom decider that log everything except successful calls from "/blah.foo.healthcheck/Check" method.
	opts := []logging.Option{
		logging.WithDecider(func(methodFullName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && methodFullName == "/blah.foo.healthcheck/Check" {
				return false
			}

			// by default you will log all calls
			return true
		}),
	}
	// Create a server, make sure we put the tags context before everything else.
	_ = []grpc.ServerOption{
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(grpclogrus.InterceptorLogger(logger), opts...),
		),
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(),
			logging.StreamServerInterceptor(grpclogrus.InterceptorLogger(logger), opts...),
		),
	}
}

func ExampleWithPayloadLogging() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := logrus.New()
	// Expect payload from  "/blah.foo.healthcheck/Check" call to be logged.
	payloadDecider := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
		return fullMethodName == "/blah.foo.healthcheck/Check"
	}

	// Create a server, make sure we put the tags context before everything else.
	_ = []grpc.ServerOption{
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(grpclogrus.InterceptorLogger(logger)),
			logging.PayloadUnaryServerInterceptor(grpclogrus.InterceptorLogger(logger), payloadDecider),
		),
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(),
			logging.StreamServerInterceptor(grpclogrus.InterceptorLogger(logger)),
			logging.PayloadStreamServerInterceptor(grpclogrus.InterceptorLogger(logger), payloadDecider),
		),
	}
}
