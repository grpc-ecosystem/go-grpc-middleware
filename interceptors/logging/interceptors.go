package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

// extractFields returns all fields from tags.
func extractFields(tags tags.Tags) Fields {
	var fields Fields
	for k, v := range tags.Values() {
		fields = append(fields, k, v)
	}
	return fields
}

type reporter struct {
	ctx             context.Context
	typ             interceptors.GRPCType
	service, method string

	requestLoggerDecider bool

	opts   *options
	logger Logger

	kind string
}

// PostCall logs the server/method name details and error if any.
func (c *reporter) PostCall(err error, duration time.Duration) {
	if !c.opts.shouldLog(interceptors.FullMethod(c.service, c.method), err) {
		return
	}
	if err == io.EOF {
		err = nil
	}
	// Get optional, fresh tags.
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)

	code := c.opts.codeFunc(err)
	logger = logger.With("grpc.code", code.String())
	if err != nil {
		logger = logger.With("error", fmt.Sprintf("%v", err))
	}
	logger.With(c.opts.durationFieldFunc(duration)...).Log(c.opts.levelFunc(code), fmt.Sprintf("finished %s %s call", c.kind, c.typ))
}

// PostMsgSend logs the details of the servingObject(req/resp) that is flowing out of the rpc.
func (c *reporter) PostMsgSend(req interface{}, err error, duration time.Duration) {

	// If the Request Logging is configured to be skipped.
	if !c.requestLoggerDecider {
		return
	}
	code := c.opts.codeFunc(err)
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logRequestResponse(c.ctx, logger.With("grpc.recv.duration", duration.String()), req, c.opts.levelFunc(code))

}

// PostMsgReceive logs the details of the servingObject that is flowing into the rpc.
func (c *reporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {

	// If the Request Logging is configured to be skipped.
	if !c.requestLoggerDecider {
		return
	}
	code := c.opts.codeFunc(err)
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logRequestResponse(c.ctx, logger.With("grpc.recv.duration", duration.String()), reply, c.opts.levelFunc(code))

}

type reportable struct {
	opts                 *options
	logger               Logger
	requestLoggerDecider PostRequestLoggingDecider
}

func (r *reportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	reqLoggerDecider := r.requestLoggerDecider(ctx, interceptors.FullMethod(service, method), req)
	return r.reporter(ctx, typ, service, method, KindServerFieldValue, reqLoggerDecider)
}

func (r *reportable) ClientReporter(ctx context.Context, resp interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	reqLoggerDecider := r.requestLoggerDecider(ctx, interceptors.FullMethod(service, method), resp)
	return r.reporter(ctx, typ, service, method, KindClientFieldValue, reqLoggerDecider)
}

func (r *reportable) reporter(ctx context.Context, typ interceptors.GRPCType, service string, method string, kind string, requestLoggerDecider bool) (interceptors.Reporter, context.Context) {

	fields := commonFields(kind, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &reporter{
		ctx:                  ctx,
		typ:                  typ,
		service:              service,
		method:               method,
		requestLoggerDecider: requestLoggerDecider,
		opts:                 r.opts,
		logger:               r.logger.With(fields...),
		kind:                 kind,
	}, ctx
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from tags package) available in current context as fields.
func UnaryClientInterceptor(logger Logger, decider PostRequestLoggingDecider, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.UnaryClientInterceptor(&reportable{logger: logger, requestLoggerDecider: decider, opts: o})
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from tags package) available in current context as fields.
func StreamClientInterceptor(logger Logger, decider PostRequestLoggingDecider, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.StreamClientInterceptor(&reportable{logger: logger, requestLoggerDecider: decider, opts: o})
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from tags package) available in current context as fields.
func UnaryServerInterceptor(logger Logger, decider PostRequestLoggingDecider, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.UnaryServerInterceptor(&reportable{logger: logger, requestLoggerDecider: decider, opts: o})
}

// StreamServerInterceptor returns a new stream server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from tags package) available in current context as fields.
func StreamServerInterceptor(logger Logger, decider PostRequestLoggingDecider, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.StreamServerInterceptor(&reportable{logger: logger, requestLoggerDecider: decider, opts: o})
}

// TODO: yashrsharma44
// What are the things that we want to log?
func logRequestResponse(ctx context.Context, logger Logger, servObj interface{}, level Level) {

	reqId, ok := RequestIDFromContext(ctx)
	if ok {
		logger = logger.With("request-id", reqId)
	} else {
		logger = logger.With("request-id", "NONE")
	}

	logger.Log(level, "logged details of the request")
}

// RequestIDFromContext returns the request id from context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	headers, ok := metadata.FromIncomingContext(ctx)
	rid := headers["X-Request-ID"][0]
	return rid, ok
}
