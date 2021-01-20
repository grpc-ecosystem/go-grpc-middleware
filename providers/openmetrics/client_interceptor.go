package metrics

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	prom "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	// DefaultClientMetrics is the default instance of ClientMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultClientMetrics = NewClientMetrics()
)

func init() {
	prom.MustRegister(DefaultClientMetrics.clientStartedCounter)
	prom.MustRegister(DefaultClientMetrics.clientHandledCounter)
	prom.MustRegister(DefaultClientMetrics.clientStreamMsgReceived)
	prom.MustRegister(DefaultClientMetrics.clientStreamMsgSent)
}

// UnaryClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&reportable{})
}

// StreamClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&reportable{})
}

// EnableClientHandlingTimeHistogram turns on recording of handling time of
// RPCs. Histogram metrics can be very expensive for Prometheus to retain and
// query. This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientHandlingTimeHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableClientHandlingTimeHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientHandledHistogram)
}

// EnableClientStreamReceiveTimeHistogram turns on recording of
// single message receive time of streaming RPCs.
// This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientStreamReceiveTimeHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableClientStreamReceiveTimeHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientStreamRecvHistogram)
}

// EnableClientStreamSendTimeHistogram turns on recording of
// single message send time of streaming RPCs.
// This function acts on the DefaultClientMetrics variable and the
// default Prometheus metrics registry.
func EnableClientStreamSendTimeHistogram(opts ...HistogramOption) {
	DefaultClientMetrics.EnableClientStreamSendTimeHistogram(opts...)
	prom.Register(DefaultClientMetrics.clientStreamSendHistogram)
}
