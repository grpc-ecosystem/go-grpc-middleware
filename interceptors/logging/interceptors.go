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
	startCallLogged bool
	opts            *options
	logger          Logger
	kind            string
}

func (c *reporter) logMessage(logger Logger, err error, msg string, duration time.Duration) {
	code := c.opts.codeFunc(err)
	logger = logger.With("grpc.code", code.String())
	if err != nil {
		logger = logger.With("grpc.error", fmt.Sprintf("%v", err))
	}
	logger = logger.With(extractFields(tags.Extract(c.ctx))...)
	logger.With(c.opts.durationFieldFunc(duration)...).Log(c.opts.levelFunc(code), msg)
}

func (c *reporter) PostCall(err error, duration time.Duration) {
	switch c.opts.shouldLog(interceptors.FullMethod(c.service, c.method)) {
	case LogFinishCall, LogStartAndFinishCall:
		if err == io.EOF {
			err = nil
		}
		c.logMessage(c.logger, err, "finished call", duration)
	default:
		return
	}
}

func (c *reporter) PostMsgSend(_ interface{}, err error, duration time.Duration) {
	if c.startCallLogged {
		return
	}
	switch c.opts.shouldLog(interceptors.FullMethod(c.service, c.method)) {
	case LogStartAndFinishCall:
		c.startCallLogged = true
		c.logMessage(c.logger, err, "started call", duration)
	}
}

func (c *reporter) PostMsgReceive(_ interface{}, err error, duration time.Duration) {
	if c.startCallLogged {
		return
	}
	switch c.opts.shouldLog(interceptors.FullMethod(c.service, c.method)) {
	case LogStartAndFinishCall:
		c.startCallLogged = true
		c.logMessage(c.logger, err, "started call", duration)
	}
}

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
		ctx:             ctx,
		typ:             typ,
		service:         service,
		method:          method,
		startCallLogged: false,
		opts:            r.opts,
		logger:          r.logger.With(fields...),
		kind:            kind,
	}, ctx
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from tags package) available in current context as fields.
func UnaryClientInterceptor(logger Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.UnaryClientInterceptor(&reportable{logger: logger, opts: o})
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
// Logger will use all tags (from tags package) available in current context as fields.
func StreamClientInterceptor(logger Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.StreamClientInterceptor(&reportable{logger: logger, opts: o})
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from tags package) available in current context as fields.
func UnaryServerInterceptor(logger Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.UnaryServerInterceptor(&reportable{logger: logger, opts: o})
}

// StreamServerInterceptor returns a new stream server interceptors that optionally logs endpoint handling.
// Logger will use all tags (from tags package) available in current context as fields.
func StreamServerInterceptor(logger Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.StreamServerInterceptor(&reportable{logger: logger, opts: o})
}
