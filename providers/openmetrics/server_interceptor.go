package metrics

import (
	openmetrics "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

var (
	// DefaultServerMetrics is the default instance of ServerMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultServerMetrics = NewServerMetrics(openmetrics.DefaultRegisterer)
)

func init() {
	DefaultServerMetrics.MustRegister(DefaultServerMetrics.serverStartedCounter)
	DefaultServerMetrics.MustRegister(DefaultServerMetrics.serverHandledCounter)
	DefaultServerMetrics.MustRegister(DefaultServerMetrics.serverStreamMsgReceived)
	DefaultServerMetrics.MustRegister(DefaultServerMetrics.serverStreamMsgSent)
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func UnaryServerInterceptor(serverRegister openmetrics.Registerer) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&reportable{
		registry: serverRegister,
	})
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func StreamServerInterceptor(serverRegister openmetrics.Registerer) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&reportable{
		registry: serverRegister,
	})
}

// RegisterAllMetrics takes a gRPC server and pre-initializes all counters to 0. This
// allows for easier monitoring in Prometheus (no missing metrics), and should
// be called *after* all services have been registered with the server. This
// function acts on the DefaultServerMetrics variable.
// If you are using a custom registry, use customServerMetrics.IntitializeMetrics() for the same.
func RegisterAllMetrics(server *grpc.Server) {
	DefaultServerMetrics.InitializeMetrics(server)
}
