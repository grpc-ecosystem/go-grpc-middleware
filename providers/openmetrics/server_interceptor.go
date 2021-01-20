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
	DefaultServerMetrics = NewServerMetrics()
)

func init() {
	openmetrics.MustRegister(DefaultServerMetrics.serverStartedCounter)
	openmetrics.MustRegister(DefaultServerMetrics.serverHandledCounter)
	openmetrics.MustRegister(DefaultServerMetrics.serverStreamMsgReceived)
	openmetrics.MustRegister(DefaultServerMetrics.serverStreamMsgSent)
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&reportable{})
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&reportable{})
}

// Register takes a gRPC server and pre-initializes all counters to 0. This
// allows for easier monitoring in Prometheus (no missing metrics), and should
// be called *after* all services have been registered with the server. This
// function acts on the DefaultServerMetrics variable.
func Register(server *grpc.Server) {
	DefaultServerMetrics.InitializeMetrics(server)
}

// EnableHandlingTimeHistogram turns on recording of handling time
// of RPCs. Histogram metrics can be very expensive for Prometheus
// to retain and query. This function acts on the DefaultServerMetrics
// variable and the default Prometheus metrics registry.
func EnableHandlingTimeHistogram(opts ...HistogramOption) {
	DefaultServerMetrics.EnableHandlingTimeHistogram(opts...)
	openmetrics.Register(DefaultServerMetrics.serverHandledHistogram)
}
