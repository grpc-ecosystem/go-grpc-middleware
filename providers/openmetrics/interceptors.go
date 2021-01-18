package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

type reporter struct {
	clientMetrics   *ClientMetrics
	serverMetrics   *ServerMetrics
	typ             interceptors.GRPCType
	service, method string
	startTime       time.Time
	kind            Kind
}

func (r *reporter) StartTimeCall(startTime time.Time, callType string) interceptors.Timer {
	switch r.kind {
	case KindClient:
		switch callType {
		case string(interceptors.Send):
			if r.clientMetrics.clientStreamSendHistogramEnabled {
				hist := r.clientMetrics.clientStreamSendHistogram.WithLabelValues(string(r.typ), r.service, r.method)
				return prometheus.NewTimer(hist)
			}
		case string(interceptors.Receive):
			if r.clientMetrics.clientStreamRecvHistogramEnabled {
				hist := r.clientMetrics.clientStreamRecvHistogram.WithLabelValues(string(r.typ), r.service, r.method)
				return prometheus.NewTimer(hist)
			}
		}
	}
	return interceptors.EmptyTimer
}

func (r *reporter) PostCall(err error, duration time.Duration) {
	// get status code from error
	status, _ := FromError(err)
	code := status.Code()

	// perform handling of metrics from code
	switch r.kind {
	case KindServer:
		r.serverMetrics.serverHandledCounter.WithLabelValues(string(r.typ), r.service, r.method, code.String()).Inc()
		if r.serverMetrics.serverHandledHistogramEnabled {
			r.serverMetrics.serverHandledHistogram.WithLabelValues(string(r.typ), r.service, r.method).Observe(time.Since(r.startTime).Seconds())
		}
	case KindClient:
		r.clientMetrics.clientHandledCounter.WithLabelValues(string(r.typ), r.service, r.method, code.String()).Inc()
		if r.clientMetrics.clientHandledHistogramEnabled {
			r.clientMetrics.clientHandledHistogram.WithLabelValues(string(r.typ), r.service, r.method).Observe(time.Since(r.startTime).Seconds())
		}
	}
}

func (r *reporter) PostMsgSend(_ interface{}, _ error, _ time.Duration) {
	switch r.kind {
	case KindServer:
		r.serverMetrics.serverStreamMsgSent.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	case KindClient:
		r.clientMetrics.clientStreamMsgSent.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	}
}

func (r *reporter) PostMsgReceive(_ interface{}, _ error, _ time.Duration) {
	switch r.kind {
	case KindServer:
		r.serverMetrics.serverStreamMsgReceived.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	case KindClient:
		r.clientMetrics.clientStreamMsgReceived.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	}
}

type reportable struct {
}

func (rep *reportable) ServerReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	m := NewServerMetrics()
	return rep.reporter(m, nil, typ, service, method, KindServer)
}

func (rep *reportable) ClientReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	m := NewClientMetrics()
	return rep.reporter(nil, m, typ, service, method, KindClient)
}

func (rep *reportable) reporter(sm *ServerMetrics, cm *ClientMetrics, rpcType interceptors.GRPCType, service, method string, kind Kind) (interceptors.Reporter, context.Context) {
	r := &reporter{
		clientMetrics: cm,
		serverMetrics: sm,
		typ:           rpcType,
		service:       service,
		method:        method,
		kind:          kind,
	}

	switch kind {
	case KindClient:
		if r.clientMetrics.clientHandledHistogramEnabled {
			r.startTime = time.Now()
		}
		r.clientMetrics.clientStartedCounter.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	case KindServer:
		if r.serverMetrics.serverHandledHistogramEnabled {
			r.startTime = time.Now()
		}
		r.serverMetrics.serverStartedCounter.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	}

	// TODO: @yashrsharma - What should we instead of the context interface?
	return r, context.Background()
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&reportable{})
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&reportable{})
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&reportable{})
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&reportable{})
}
