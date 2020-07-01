package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"

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

	// The error if any would be handled by the PostCall
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logRequestResponse(logger)

}

// PostMsgReceive logs the details of the servingObject that is flowing into the rpc.
func (c *reporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {

	// The error if any would be handled by the PostCall
	if err != nil {
		return
	}

	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logRequestResponse(logger)

}

type reportable struct {
	clientDecider ClientRequestLoggingDecider
	serverDecider ServerRequestLoggingDecider
	opts          *options
	logger        Logger
}

func (r *reportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {

	if !r.serverDecider(ctx, interceptors.FullMethod(service, method), req) {
		return interceptors.NoopReporter{}, ctx
	}

	return r.reporter(ctx, typ, service, method, KindServerFieldValue)
}

func (r *reportable) ClientReporter(ctx context.Context, resp interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.clientDecider(ctx, interceptors.FullMethod(service, method), resp) {
		return interceptors.NoopReporter{}, ctx
	}

	return r.reporter(ctx, typ, service, method, KindClientFieldValue)
}

func (r *reportable) reporter(ctx context.Context, typ interceptors.GRPCType, service string, method string, kind string) (interceptors.Reporter, context.Context) {
	fields := commonFields(kind, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &reporter{
		ctx:     ctx,
		typ:     typ,
		service: service,
		method:  method,
		opts:    r.opts,
		logger:  r.logger.With(fields...),
		kind:    kind,
	}, ctx
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from tags package) available in current context as fields.
func UnaryClientInterceptor(logger Logger, decider ClientRequestLoggingDecider, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.UnaryClientInterceptor(&reportable{logger: logger, clientDecider: decider, opts: o})
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from tags package) available in current context as fields.
func StreamClientInterceptor(logger Logger, decider ClientRequestLoggingDecider, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.StreamClientInterceptor(&reportable{logger: logger, clientDecider: decider, opts: o})
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from tags package) available in current context as fields.
func UnaryServerInterceptor(logger Logger, decider ServerRequestLoggingDecider, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.UnaryServerInterceptor(&reportable{logger: logger, serverDecider: decider, opts: o})
}

// StreamServerInterceptor returns a new stream server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from tags package) available in current context as fields.
func StreamServerInterceptor(logger Logger, decider ServerRequestLoggingDecider, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.StreamServerInterceptor(&reportable{logger: logger, serverDecider: decider, opts: o})
}

// TODO: yashrsharma44
// What are the things that we want to log?
func logRequestResponse(Logger) {}
