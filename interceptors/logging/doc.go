// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

/*
Package logging is a "parent" package for gRPC logging middlewares.

The gRPC logging middleware populates request-scoped data to `logging.Fields` that relate to the current gRPC call
(e.g. service and method names). You can extract/inject data in the propagated context using `logging.ExtractFields` and `logging.InjectFields`.

Once the gRPC logging middleware has added the gRPC specific Fields to the ctx they will then be written with the log lines.

All logging middleware will emit a final log statement. It is based on the error returned by the handler function,
the gRPC status code, an error (if any) and it emits at a level controlled via `WithLevels`. You can control this behavior
using `WithDecider`.

# Fields discovered by a handler

To include a field discovered by an RPC handler in the interceptor's finish-call log, call [AddFields] on the handler's original context before returning, with [FinishCall] logging enabled:

	func (s *server) GetWidget(ctx context.Context, req *GetWidgetRequest) (*GetWidgetResponse, error) {
		logging.AddFields(ctx, logging.Fields{"widget.id", req.GetId()})
		// Handle the request.
	}

[InjectFields] instead returns a derived context. It is useful for passing fields to downstream code, but
assigning it to ctx inside a handler does not change the context held by the logging interceptor.

For streaming handlers, pass the stream's Context to [AddFields]. Call it only when no other goroutine can
read or update the fields, including through concurrent stream operations that log messages. Fields added
by a handler do not change log entries that have already been emitted.

See the Zap example (examples/zap) for a runnable unary RPC that adds a field to the finish-call record.

# This parent package

This particular package is intended for use by other middleware, logging or otherwise. It contains interfaces that other
logging middlewares *could* share. This allows code to be shared between different implementations.

# Field names

All field names of loggers follow the OpenTracing semantics definitions, with `grpc.` prefix if needed:
https://github.com/opentracing/specification/blob/master/semantic_conventions.md

Implementations:

* providers/kit
* providers/logr
* providers/logrus
* providers/phuslog
* providers/slog
* providers/zap
* providers/zerolog
*/
package logging
