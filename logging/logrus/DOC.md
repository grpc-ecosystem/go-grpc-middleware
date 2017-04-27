# grpc_logrus
--
    import "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"

`grpc_logrus` is a gRPC logging middleware backed by Logrus loggers

It accepts a user-configured `logrus.Entry` that will be used for logging
completed gRPC calls. The same `logrus.Entry` will be used for logging completed
gRPC calls, and be populated into the `context.Context` passed into gRPC handler
code.

You can use `Extract` to log into a request-scoped `logrus.Entry` instance in
your handler code. The fields set on the logger correspond to the
grpc_ctxtags.Tags attached to the context.

Logrus can also be made as a backend for gRPC library internals. For that use
`ReplaceGrpcLogger`.

Please see examples and tests for examples of use.

## Usage

```go
var (
	// SystemField is used in every log statement made through grpc_logrus. Can be overwritten before any initialization code.
	SystemField = "system"

	// KindField describes the log gield used to incicate whether this is a server or a client log statment.
	KindField = "span.kind"
)
```

```go
var DefaultDurationToField = DurationToTimeMillisField
```
DefaultDurationToField is the default implementation of converting request
duration to a log field (key and value).

```go
var (
	// JsonPBMarshaller is the marshaller used for serializing protobuf messages.
	JsonPbMarshaller = &jsonpb.Marshaler{}
)
```

#### func  DefaultClientCodeToLevel

```go
func DefaultClientCodeToLevel(code codes.Code) logrus.Level
```
DefaultClientCodeToLevel is the default implementation of gRPC return codes to
log levels for client side.

#### func  DefaultCodeToLevel

```go
func DefaultCodeToLevel(code codes.Code) logrus.Level
```
DefaultCodeToLevel is the default implementation of gRPC return codes to log
levels for server side.

#### func  DurationToDurationField

```go
func DurationToDurationField(duration time.Duration) (key string, value interface{})
```
DurationToDurationField uses the duration value to log the request duration.

#### func  DurationToTimeMillisField

```go
func DurationToTimeMillisField(duration time.Duration) (key string, value interface{})
```
DurationToTimeMillisField converts the duration to milliseconds and uses the key
`grpc.time_ms`.

#### func  Extract

```go
func Extract(ctx context.Context) *logrus.Entry
```
Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.

If the grpc_logrus middleware wasn't used, a no-op `logrus.Entry` is returned.
This makes it safe to use regardless.

#### func  PayloadStreamClientInterceptor

```go
func PayloadStreamClientInterceptor(entry *logrus.Entry, decider grpc_logging.ClientPayloadLoggingDecider) grpc.StreamClientInterceptor
```
PayloadStreamServerInterceptor returns a new streaming client interceptor that
logs the paylods of requests and responses.

#### func  PayloadStreamServerInterceptor

```go
func PayloadStreamServerInterceptor(entry *logrus.Entry, decider grpc_logging.ServerPayloadLoggingDecider) grpc.StreamServerInterceptor
```
PayloadUnaryServerInterceptor returns a new server server interceptors that logs
the payloads of requests.

This *only* works when placed *after* the `grpc_logrus.StreamServerInterceptor`.
However, the logging can be done to a separate instance of the logger.

#### func  PayloadUnaryClientInterceptor

```go
func PayloadUnaryClientInterceptor(entry *logrus.Entry, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor
```
PayloadUnaryClientInterceptor returns a new unary client interceptor that logs
the paylods of requests and responses.

#### func  PayloadUnaryServerInterceptor

```go
func PayloadUnaryServerInterceptor(entry *logrus.Entry, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor
```
PayloadUnaryServerInterceptor returns a new unary server interceptors that logs
the payloads of requests.

This *only* works when placed *after* the `grpc_logrus.UnaryServerInterceptor`.
However, the logging can be done to a separate instance of the logger.

#### func  ReplaceGrpcLogger

```go
func ReplaceGrpcLogger(logger *logrus.Entry)
```
ReplaceGrpcLogger sets the given logrus.Logger as a gRPC-level logger. This
should be called *before* any other initialization, preferably from init()
functions.

#### func  StreamClientInterceptor

```go
func StreamClientInterceptor(entry *logrus.Entry, opts ...Option) grpc.StreamClientInterceptor
```
StreamServerInterceptor returns a new streaming client interceptor that
optionally logs the execution of external gRPC calls.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that adds
logrus.Entry to the context.

#### func  UnaryClientInterceptor

```go
func UnaryClientInterceptor(entry *logrus.Entry, opts ...Option) grpc.UnaryClientInterceptor
```
UnaryClientInterceptor returns a new unary client interceptor that optionally
logs the execution of external gRPC calls.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.UnaryServerInterceptor
```
PayloadUnaryServerInterceptor returns a new unary server interceptors that adds
logrus.Entry to the context.

#### type CodeToLevel

```go
type CodeToLevel func(code codes.Code) logrus.Level
```

CodeToLevel function defines the mapping between gRPC return codes and
interceptor log level.

#### type DurationToField

```go
type DurationToField func(duration time.Duration) (key string, value interface{})
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
WithDurationField customizes the function for mapping request durations to log
fields.

#### func  WithLevels

```go
func WithLevels(f CodeToLevel) Option
```
WithLevels customizes the function for mapping gRPC return codes and interceptor
log level statements.
