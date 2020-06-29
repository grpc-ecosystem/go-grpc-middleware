package logging

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"google.golang.org/grpc"
)

type serverRequestReporter struct {
	ctx    context.Context
	logger Logger
}

// Log the server/method name details and error if any
func (s *serverRequestReporter) PostCall(error, time.Duration) {}

// Log the details of the resp that is flowing out of the server
func (s *serverRequestReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {

	// The error if any, would be handled by the PostCall
	if err != nil {
		return
	}
	logger := s.logger.With(extractFields(tags.Extract(s.ctx))...)
	logRequestResponse(logger)
}

// Log the details of the req that is flowing into it
func (s *serverRequestReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {

	// The error if any would be handled by the PostCall
	if err != nil {
		return
	}

	logger := s.logger.With(extractFields(tags.Extract(s.ctx))...)
	logRequestResponse(logger)

}

type clientRequestReporter struct {
	ctx    context.Context
	logger Logger
}

func (c *clientRequestReporter) PostCall(error, time.Duration) {}

func (c *clientRequestReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {

	// The error if any would be handled by the PostCall
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logRequestResponse(logger)

}

func (c *clientRequestReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {

	// The error if any would be handled by the PostCall
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logRequestResponse(logger)
}

type requestReportable struct {
	clientDecider ClientRequestLoggingDecider
	serverDecider ServerRequestLoggingDecider
	logger        Logger
}

func (r *requestReportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {

	if !r.serverDecider(ctx, interceptors.FullMethod(service, method), req) {
		return interceptors.NoopReporter{}, ctx
	}

	fields := commonFields(KindServerFieldValue, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &serverRequestReporter{
		ctx:    ctx,
		logger: r.logger.With(fields...),
	}, ctx

}

func (r *requestReportable) ClientReporter(ctx context.Context, resp interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.clientDecider(ctx, interceptors.FullMethod(service, method), resp) {
		return interceptors.NoopReporter{}, ctx
	}
	fields := commonFields(KindClientFieldValue, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &clientRequestReporter{
		ctx:    ctx,
		logger: r.logger.With(fields...),
	}, ctx
}

// RequestUnaryServerInterceptor returns a new unary server interceptors that logs the metadata of requests on INFO level.
// Logger tags will be used from tags context.
func RequestUnaryServerInterceptor(logger Logger, decider ServerRequestLoggingDecider) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&requestReportable{logger: logger, serverDecider: decider})
}

// RequestStreamServerInterceptor returns a new server server interceptors that logs the metadata of requests on INFO level.
// Logger tags will be used from tags context.
func RequestStreamServerInterceptor(logger Logger, decider ServerRequestLoggingDecider) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&requestReportable{logger: logger, serverDecider: decider})
}

// RequestUnaryClientInterceptor returns a new unary client interceptor that logs the metadata of requests and responses on INFO level.
// Logger tags will be used from tags context.
func RequestUnaryClientInterceptor(logger Logger, decider ClientRequestLoggingDecider) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&requestReportable{logger: logger, clientDecider: decider})
}

// RequestStreamClientInterceptor returns a new streaming client interceptor that logs the metadata of requests and responses on INFO level.
// Logger tags will be used from tags context.
func RequestStreamClientInterceptor(logger Logger, decider ClientRequestLoggingDecider) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&requestReportable{logger: logger, clientDecider: decider})
}

// TODO: yashrsharma44
// What are the things that we want to log?
func logRequestResponse(Logger) {}
