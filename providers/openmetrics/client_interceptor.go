package metrics

import (
	openmetrics "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

var (
	// DefaultClientMetrics is the default instance of ClientMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultClientMetrics = NewClientMetrics(openmetrics.DefaultRegisterer)
)

func init() {
	DefaultClientMetrics.MustRegister(DefaultClientMetrics.clientStartedCounter)
	DefaultClientMetrics.MustRegister(DefaultClientMetrics.clientHandledCounter)
	DefaultClientMetrics.MustRegister(DefaultClientMetrics.clientStreamMsgReceived)
	DefaultClientMetrics.MustRegister(DefaultClientMetrics.clientStreamMsgSent)
}

// UnaryClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
func UnaryClientInterceptor(clientRegister openmetrics.Registerer) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&reportable{
		registry: clientRegister,
	})
}

// StreamClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func StreamClientInterceptor(clientRegister openmetrics.Registerer) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&reportable{
		registry: clientRegister,
	})
}
