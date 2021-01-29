package metrics

import (
	"context"
	"time"

	openmetrics "github.com/prometheus/client_golang/prometheus"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
)

type reporter struct {
	clientMetrics           *ClientMetrics
	serverMetrics           *ServerMetrics
	typ                     interceptors.GRPCType
	service, method         string
	startTime               time.Time
	kind                    Kind
	sendTimer, receiveTimer interceptors.Timer
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
		r.sendTimer.ObserveDuration()
	}
}

func (r *reporter) PostMsgReceive(_ interface{}, _ error, _ time.Duration) {
	switch r.kind {
	case KindServer:
		r.serverMetrics.serverStreamMsgReceived.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	case KindClient:
		r.clientMetrics.clientStreamMsgReceived.WithLabelValues(string(r.typ), r.service, r.method).Inc()
		r.receiveTimer.ObserveDuration()
	}
}

type reportable struct {
	registry openmetrics.Registerer
}

func (rep *reportable) ServerReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	m := NewServerMetrics(rep.registry)
	return rep.reporter(m, nil, typ, service, method, KindServer)
}

func (rep *reportable) ClientReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	m := NewClientMetrics(rep.registry)
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
		sendTimer:     interceptors.EmptyTimer,
		receiveTimer:  interceptors.EmptyTimer,
	}

	switch kind {
	case KindClient:
		if r.clientMetrics.clientHandledHistogramEnabled {
			r.startTime = time.Now()
		}
		r.clientMetrics.clientStartedCounter.WithLabelValues(string(r.typ), r.service, r.method).Inc()

		if r.clientMetrics.clientStreamSendHistogramEnabled {
			hist := r.clientMetrics.clientStreamSendHistogram.WithLabelValues(string(r.typ), r.service, r.method)
			r.sendTimer = openmetrics.NewTimer(hist)
		}

		if r.clientMetrics.clientStreamRecvHistogramEnabled {
			hist := r.clientMetrics.clientStreamRecvHistogram.WithLabelValues(string(r.typ), r.service, r.method)
			r.receiveTimer = openmetrics.NewTimer(hist)
		}
	case KindServer:
		if r.serverMetrics.serverHandledHistogramEnabled {
			r.startTime = time.Now()
		}
		r.serverMetrics.serverStartedCounter.WithLabelValues(string(r.typ), r.service, r.method).Inc()
	}

	// TODO: @yashrsharma44 - What should we use instead of the context.Background()?
	return r, context.Background()
}
