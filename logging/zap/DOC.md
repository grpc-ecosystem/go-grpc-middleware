# grpc_zap
--
    import "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

`grpc_zap` is a gRPC logging middleware backed by ZAP loggers

It accepts a user-configured `zap.Logger` that will be used for logging
completed gRPC calls. The same `zap.Logger` will be used for logging completed
gRPC calls, and be populated into the `context.Context` passed into gRPC handler
code.

You can use `Extract` to log into a request-scoped `zap.Logger` instance in your
handler code. The fields set on the logger correspond to the grpc_ctxtags.Tags
attached to the context.

ZAP can also be made as a backend for gRPC library internals. For that use
`ReplaceGrpcLogger`.

Please see examples and tests for examples of use.

## Usage

```go
var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")

	// ServerField is used in every server-side log statment made through grpc_zap.Can be overwritten before initialization.
	ServerField = zap.String("span.kind", "server")
)
```

```go
var (
	// ClientField is used in every client-side log statement made through grpc_zap. Can be overwritten before initialization.
	ClientField = zap.String("span.kind", "client")
)
```

```go
var DefaultDurationToField = DurationToTimeMillisField
```
DefaultDurationToField is the default implementation of converting request
duration to a Zap field.

```go
var (
	// JsonPBMarshaller is the marshaller used for serializing protobuf messages.
	JsonPbMarshaller = &jsonpb.Marshaler{}
)
```

#### func  DefaultClientCodeToLevel

```go
func DefaultClientCodeToLevel(code codes.Code) zapcore.Level
```
DefaultClientCodeToLevel is the default implementation of gRPC return codes to
log levels for client side.

#### func  DefaultCodeToLevel

```go
func DefaultCodeToLevel(code codes.Code) zapcore.Level
```
DefaultCodeToLevel is the default implementation of gRPC return codes and
interceptor log level for server side.

#### func  DurationToDurationField

```go
func DurationToDurationField(duration time.Duration) zapcore.Field
```
DurationToDurationField uses a Duration field to log the request duration and
leaves it up to Zap's encoder settings to determine how that is output.

#### func  DurationToTimeMillisField

```go
func DurationToTimeMillisField(duration time.Duration) zapcore.Field
```
DurationToTimeMillisField converts the duration to milliseconds and uses the key
`grpc.time_ms`.

#### func  Extract

```go
func Extract(ctx context.Context) *zap.Logger
```
Extract takes the call-scoped Logger from grpc_zap middleware.

It always returns a Logger that has all the grpc_ctxtags updated.

#### func  PayloadStreamClientInterceptor

```go
func PayloadStreamClientInterceptor(logger *zap.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.StreamClientInterceptor
```
PayloadStreamServerInterceptor returns a new streaming client interceptor that
logs the paylods of requests and responses.

#### func  PayloadStreamServerInterceptor

```go
func PayloadStreamServerInterceptor(logger *zap.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.StreamServerInterceptor
```
PayloadUnaryServerInterceptor returns a new server server interceptors that logs
the payloads of requests.

This *only* works when placed *after* the `grpc_logrus.StreamServerInterceptor`.
However, the logging can be done to a separate instance of the logger.

#### func  PayloadUnaryClientInterceptor

```go
func PayloadUnaryClientInterceptor(logger *zap.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor
```
PayloadUnaryClientInterceptor returns a new unary client interceptor that logs
the paylods of requests and responses.

#### func  PayloadUnaryServerInterceptor

```go
func PayloadUnaryServerInterceptor(logger *zap.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor
```
PayloadUnaryServerInterceptor returns a new unary server interceptors that logs
the payloads of requests.

This *only* works when placed *after* the `grpc_logrus.UnaryServerInterceptor`.
However, the logging can be done to a separate instance of the logger.

#### func  ReplaceGrpcLogger

```go
func ReplaceGrpcLogger(logger *zap.Logger)
```
ReplaceGrpcLogger sets the given zap.Logger as a gRPC-level logger. This should
be called *before* any other initialization, preferably from init() functions.

#### func  StreamClientInterceptor

```go
func StreamClientInterceptor(logger *zap.Logger, opts ...Option) grpc.StreamClientInterceptor
```
StreamServerInterceptor returns a new streaming client interceptor that
optionally logs the execution of external gRPC calls.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(logger *zap.Logger, opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that adds
zap.Logger to the context.

#### func  UnaryClientInterceptor

```go
func UnaryClientInterceptor(logger *zap.Logger, opts ...Option) grpc.UnaryClientInterceptor
```
UnaryClientInterceptor returns a new unary client interceptor that optionally
logs the execution of external gRPC calls.

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

#### type DurationToField

```go
type DurationToField func(duration time.Duration) zapcore.Field
```

DurationToField function defines how to produce duration fields for logging

#### type Option

```go
type Option func(*options)
```


#### func  WithCodes

```go
func WithCodes(f grpc_logging.ErrorToCode) Option
```
WithCodes customizes the function for mapping errors to error codes.

#### func  WithDurationField

```go
func WithDurationField(f DurationToField) Option
```
WithDurationField customizes the function for mapping request durations to Zap
fields.

#### func  WithLevels

```go
func WithLevels(f CodeToLevel) Option
```
WithLevels customizes the function for mapping gRPC return codes and interceptor
log level statements.
