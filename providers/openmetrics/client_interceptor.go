package metrics

import (
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

// UnaryClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
func UnaryClientInterceptor(clientMetrics *ClientMetrics) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&reportable{
		clientMetrics: clientMetrics,
	})
}

// StreamClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func StreamClientInterceptor(clientMetrics *ClientMetrics) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&reportable{
		clientMetrics: clientMetrics,
	})
}
