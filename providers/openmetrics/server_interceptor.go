package metrics

import (
	openmetrics "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

// RegisterServerMetrics returns a custom ServerMetrics object registered
// with the user's registry, and registers some common metrics associated
// with every instance.
func RegisterServerMetrics(registry openmetrics.Registerer) *ServerMetrics {
	customServerMetrics := NewServerMetrics(registry)
	customServerMetrics.MustRegister(customServerMetrics.serverStartedCounter)
	customServerMetrics.MustRegister(customServerMetrics.serverHandledCounter)
	customServerMetrics.MustRegister(customServerMetrics.serverStreamMsgReceived)
	customServerMetrics.MustRegister(customServerMetrics.serverStreamMsgSent)

	return customServerMetrics
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
