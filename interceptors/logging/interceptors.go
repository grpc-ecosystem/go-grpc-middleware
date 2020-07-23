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
	ctx                        context.Context
	typ                        interceptors.GRPCType
	service, method            string
	firstRequestMessageLogged  bool
	firstResponseMessageLogged bool
	isServer                   bool
	opts                       *options
	logger                     Logger

	kind string
}

func (c *reporter) addMetaDetails(logger Logger) Logger {
	if c.isServer {
		logger = logger.With("component", "server")
	}
	if !c.isServer {
		logger = logger.With("component", "client")
	}
	// Attach the kind, protocol and type as a structured log entry.
	logger = logger.With("kind", c.kind)
	logger = logger.With("type", string(c.typ))
	logger = logger.With("protocol", "grpc")

	return logger
}

// PostCall logs the server/method name details and error if any.
func (c *reporter) PostCall(err error, duration time.Duration) {
	if !c.opts.shouldLog(interceptors.FullMethod(c.service, c.method), err) {
		return
	}
	if err == io.EOF {
		err = nil
	}
	logger := c.addMetaDetails(c.logger)
	// Get optional, fresh tags.
	logger = logger.With(extractFields(tags.Extract(c.ctx))...)
	code := c.opts.codeFunc(err)
	logger = logger.With("grpc.code", code.String())
	if err != nil {
		logger = logger.With("grpc.error", fmt.Sprintf("%v", err))
	}
	logger = logger.With(c.opts.durationFieldFunc(duration)...)
	logger.Log(c.opts.levelFunc(code), "finished call")
}

// PostMsgSend logs the details of the servingObject that is flowing out of the rpc.
// resp object wrt server.
// Log the details of the first request, skip if the object is response.
func (c *reporter) PostMsgSend(resp interface{}, err error, duration time.Duration) {
	if !c.opts.shouldLog(interceptors.FullMethod(c.service, c.method), err) {
		return
	}

	// If the first message is logged skip the rest of the logging.
	// If the serving object is response, skip the logging.
	if c.firstResponseMessageLogged || !c.isServer {
		return
	}
	c.firstResponseMessageLogged = true
	logger := c.addMetaDetails(c.logger)
	logger = logger.With(extractFields(tags.Extract(c.ctx))...)
	logger = logger.With("grpc.recv.duration", duration.String())
	code := c.opts.codeFunc(err)
	if err != nil {
		logger = logger.With("grpc.error", fmt.Sprintf("%v", err))
	}
	logger.With(c.opts.durationFieldFunc(duration)...).Log(c.opts.levelFunc(code), "started call")
}

// PostMsgReceive logs the details of the servingObject that is flowing into the rpc.
// req object wrt server.
// Log the details of the request, skip if the object is response.
func (c *reporter) PostMsgReceive(req interface{}, err error, duration time.Duration) {
	if !c.opts.shouldLog(interceptors.FullMethod(c.service, c.method), err) {
		return
	}

	// If the first message for request is logged, skip the rest of the logging.
	// If the serving object is a response, skip the logging.
	if c.firstRequestMessageLogged || !c.isServer {
		return
	}
	c.firstRequestMessageLogged = true
	logger := c.addMetaDetails(c.logger)
	logger = logger.With(extractFields(tags.Extract(c.ctx))...)
	logger = logger.With("grpc.recv.duration", duration.String())
	code := c.opts.codeFunc(err)

	if err != nil {
		logger = logger.With("grpc.error", fmt.Sprintf("%v", err))
	}
	logger.With(c.opts.durationFieldFunc(duration)...).Log(c.opts.levelFunc(code), "started call")
}

type reportable struct {
	opts   *options
	logger Logger
}

func (r *reportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	return r.reporter(ctx, typ, service, method, KindServerFieldValue, true)
}

func (r *reportable) ClientReporter(ctx context.Context, resp interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	return r.reporter(ctx, typ, service, method, KindClientFieldValue, false)
}

func (r *reportable) reporter(ctx context.Context, typ interceptors.GRPCType, service string, method string, kind string, isServer bool) (interceptors.Reporter, context.Context) {
	fields := commonFields(kind, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &reporter{
		ctx:                        ctx,
		typ:                        typ,
		service:                    service,
		method:                     method,
		firstRequestMessageLogged:  false,
		firstResponseMessageLogged: false,
		isServer:                   isServer,
		opts:                       r.opts,
		logger:                     r.logger.With(fields...),
		kind:                       kind,
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
