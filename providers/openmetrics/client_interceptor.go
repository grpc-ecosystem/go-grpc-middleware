package metrics

import (
	openmetrics "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

// RegisterClientMetrics returns a custom ClientMetrics object registered
// with the user's registry, and registers some common metrics associated
// with every instance.
func RegisterClientMetrics(registry openmetrics.Registerer) *ClientMetrics {
	customClientMetrics := NewClientMetrics(registry)
	customClientMetrics.MustRegister(customClientMetrics.clientStartedCounter)
	customClientMetrics.MustRegister(customClientMetrics.clientHandledCounter)
	customClientMetrics.MustRegister(customClientMetrics.clientStreamMsgReceived)
	customClientMetrics.MustRegister(customClientMetrics.clientStreamMsgSent)

	return customClientMetrics
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
