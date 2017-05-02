# grpc_logging
--
    import "github.com/grpc-ecosystem/go-grpc-middleware/logging"

grpc_logging is a "parent" package for gRPC logging middlewares


### General functionality of all middleware

All logging middleware have an `Extract(ctx)` function that provides a
request-scoped logger with gRPC-related fields (service and method names).
Moreover, that logger will have fields populated from the `grpc_ctxtags.Tags` of
the context.

All logging middleware will emit a final log statement. It is based on the error
returned by the handler function, the gRPC status code, an error (if any) and it
will emit at a level controlled via `WithLevels`.


### This parent package

This particular package is intended for use by other middleware, logging or
otherwise. It contains interfaces that other logging middlewares *could* share .
This allows code to be shared between different implementations.


### Field names

All field names of loggers follow the OpenTracing semantics definitions, with
`grpc.` prefix if needed:
https://github.com/opentracing/specification/blob/master/semantic_conventions.md


### Implementations

There are two implementations at the moment: logrus and zap

See relevant packages below.

## Usage

#### func  DefaultErrorToCode

```go
func DefaultErrorToCode(err error) codes.Code
```

#### type ClientPayloadLoggingDecider

```go
type ClientPayloadLoggingDecider func(ctx context.Context, fullMethodName string) bool
```

ClientPayloadLoggingDecider is a user-provided function for deciding whether to
log the client-side request/response payloads

#### type ErrorToCode

```go
type ErrorToCode func(err error) codes.Code
```

ErrorToCode function determines the error code of an error This makes using
custom errors with grpc middleware easier

#### type ServerPayloadLoggingDecider

```go
type ServerPayloadLoggingDecider func(ctx context.Context, fullMethodName string, servingObject interface{}) bool
```

ServerPayloadLoggingDecider is a user-provided function for deciding whether to
log the server-side request/response payloads
