# grpc_logrus
--
    import "github.com/mwitkow/go-grpc-middleware/logging/logrus"

Logrus loggers for gRPC.

`grpc_logrus` provides server-side interceptors and handler funcitions for using
Logrus within gRPC code.

It accepts a user-configured `logrus.Entry` object that is:

    * used for logging completed gRPC calls (method called, time elapsed, error code and message, log level)
    * populated into the `context.Context` passed into gRPC handler code.

You can use `Extract` to log into a request-scoped `logrus.Entry` instance in
your handler code. Moreover you `AddFields` to the request-scoped
`logrus.Entry`, that will be propagated for all call depending on the context,
including the interceptor's own "finished RPC" log message.

### The latter is very useful when the handler code wants to add additional metadata
to the call after extracting it from the request. For use cases when a
"downstream" interceptor needs to log something, please consider using
grpc_commonlog library.

### To make sure that Logrus is also receiving the log statements from the gRPC
library internals, please call `ReplaceGrpcLogger`.

Please see examples and tests for examples of use.

## Usage

```go
var (
	// SystemField is used in every log statement made through grpc_logrus. Can be overwritten before any initialization code.
	SystemField = "system"
)
```

#### func  AddFields

```go
func AddFields(ctx context.Context, fields logrus.Fields)
```
AddFields adds logrus.Fields to *all* usages of the logger, both upstream (to
handler) and downstream (to interceptor).

This call *is not* concurrency safe. It should only be used in the request
goroutine: in other interceptors or directly in the handler.

#### func  DefaultCodeToLevel

```go
func DefaultCodeToLevel(code codes.Code) logrus.Level
```
DefaultCodeToLevel is the default implementation of gRPC return codes and
interceptor log level.

#### func  Extract

```go
func Extract(ctx context.Context) *logrus.Entry
```
Extract takes the call-scoped logrus.Entry from grpc_logrus middleware.

If the grpc_logrus middleware wasn't used, a no-op `logrus.Entry` is returned.
This makes it safe to use regardless.

#### func  ReplaceGrpcLogger

```go
func ReplaceGrpcLogger(logger *logrus.Entry)
```
ReplaceGrpcLogger sets the given logrus.Logger as a gRPC-level logger. This
should be called *before* any other initialization, preferably from init()
functions.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that adds
logrus.Entry to the context.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptors that adds
logrus.Entry to the context.

#### type CodeToLevel

```go
type CodeToLevel func(code codes.Code) logrus.Level
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
