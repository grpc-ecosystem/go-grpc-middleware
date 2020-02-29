package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/interceptors"
	ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/interceptors/tags"
	"google.golang.org/grpc"
)

// extractFields returns all fields from tags.
func extractFields(tags ctxtags.Tags) Fields {
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

func (c *reporter) PostCall(err error, duration time.Duration) {
	if !c.opts.shouldLog(interceptors.FullMethod(c.service, c.method), err) {
		return
	}
	if err == io.EOF {
		err = nil
	}
	// Get optional, fresh tags.
	logger := c.logger.With(extractFields(ctxtags.Extract(c.ctx))...)

	code := c.opts.codeFunc(err)
	logger = logger.With("grpc.code", code.String())
	if err != nil {
		logger = logger.With("error", fmt.Sprintf("%v", err))
	}
	logger.With(c.opts.durationFieldFunc(duration)...).Log(c.opts.levelFunc(code), fmt.Sprintf("finished %s %s call", c.kind, c.typ))
}

func (c *reporter) PostMsgSend(interface{}, error, time.Duration) {}

func (c *reporter) PostMsgReceive(interface{}, error, time.Duration) {}

type reportable struct {
	opts   *options
	logger Logger
}

func (r *reportable) ServerReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	return r.reporter(ctx, typ, service, method, KindServerFieldValue)
}

func (r *reportable) ClientReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
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
// Logger will use all tags (from ctxtags package) available in current context as fields.
func UnaryClientInterceptor(logger Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.UnaryClientInterceptor(&reportable{logger: logger, opts: o})
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from ctxtags package) available in current context as fields.
func StreamClientInterceptor(logger Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.StreamClientInterceptor(&reportable{logger: logger, opts: o})
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from ctxtags package) available in current context as fields.
func UnaryServerInterceptor(logger Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.UnaryServerInterceptor(&reportable{logger: logger, opts: o})
}

// StreamServerInterceptor returns a new stream server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from ctxtags package) available in current context as fields.
func StreamServerInterceptor(logger Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.StreamServerInterceptor(&reportable{logger: logger, opts: o})
}
