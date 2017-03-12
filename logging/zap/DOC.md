# grpc_zap
--
    import "github.com/mwitkow/go-grpc-middleware/logging/zap"

`grpc_zap` is a gRPC logging middleware backed by ZAP loggers

It accepts a user-configured `zap.Logger` that will be used for logging
completed gRPC calls. The same `zap.Logger` will be used for logging completed
gRPC calls, and be populated into the `context.Context` passed into gRPC handler
code.

You can use `Extract` to log into a request-scoped `zap.Logger` instance in your
handler code. `AddFields` adds new fields to the request-scoped `zap.Logger`.
They will be propagated for all call depending on the context, including the
interceptor's own "finished RPC" log message.

ZAP can also be made as a backend for gRPC library internals. For that use
`ReplaceGrpcLogger`.

Please see examples and tests for examples of use.

## Usage

```go
var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")
)
```

#### func  AddFields

```go
func AddFields(ctx context.Context, fields ...zapcore.Field)
```
AddFields adds zap.Fields to *all* usages of the logger, both upstream (to
handler) and downstream.

This call *is not* concurrency safe. It should only be used in the request
goroutine: in other interceptors or directly in the handler.

#### func  DefaultCodeToLevel

```go
func DefaultCodeToLevel(code codes.Code) zapcore.Level
```
DefaultCodeToLevel is the default implementation of gRPC return codes and
interceptor log level.

#### func  Extract

```go
func Extract(ctx context.Context) *zap.Logger
```
Extract takes the call-scoped Logger from grpc_zap middleware.

If the grpc_zap middleware wasn't used, a null `zap.Logger` is returned. This
makes it safe to use regardless.

#### func  ReplaceGrpcLogger

```go
func ReplaceGrpcLogger(logger *zap.Logger)
```
ReplaceGrpcLogger sets the given zap.Logger as a gRPC-level logger. This should
be called *before* any other initialization, preferably from init() functions.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(logger *zap.Logger, opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that adds
zap.Logger to the context.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(logger *zap.Logger, opts ...Option) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptors that adds
zap.Logger to the context.

#### type CodeToLevel

```go
type CodeToLevel func(code codes.Code) zapcore.Level
```

CodeToLevel function defines the mapping between gRPC return codes and
interceptor log level.

#### type Option

```go
type Option func(*options)
```


#### func  WithFieldExtractor

```go
func WithFieldExtractor(f grpc_logging.RequestLogFieldExtractorFunc) Option
```
WithFieldExtractor customizes the function for extracting log fields from
protobuf messages.

#### func  WithLevels

```go
func WithLevels(f CodeToLevel) Option
```
WithLevels customizes the function for mapping gRPC return codes and interceptor
log level statements.

#### func  WithCodes

```go
func WithCodes(f grpc_logging.ErrorToCode) Option
```
WithCodes customizes the function for mapping errors to error codes.
