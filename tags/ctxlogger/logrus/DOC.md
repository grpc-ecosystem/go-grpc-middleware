# ctxlogger_logrus
`import "github.com/grpc-ecosystem/go-grpc-middleware/tags/ctxlogger/logrus"`

* [Overview](#pkg-overview)
* [Imported Packages](#pkg-imports)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-overview">Overview</a>
`ctxlogger_logrus` is a ctxlogger that is backed by logrus

It accepts a user-configured `logrus.Logger` that will be used for logging. The same `logrus.Logger` will
be populated into the `context.Context` passed into gRPC handler code.

On calling `StreamServerInterceptor` or `UnaryServerInterceptor` this ctxlogger middleware will add the Tags from
the ctx so that it will be present on subsequent use of the `ctxlogger_logrus` logger.

You can use `ctxlogger_zap.Extract` to log into a request-scoped `zap.Logger` instance in your handler code.

As `ctxlogger_zap.Extract` will iterate all tags on from `grpc_ctxtags` it is therefore expensive so it is advised that you
extract once at the start of the function from the context and reuse it for the remainder of the function (see examples).

Please see examples and tests for examples of use.

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
x := func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
    // Add fields the ctxtags of the request which will be added to all extracted loggers.
    grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
    // Extract a single request-scoped logrus.Logger and log messages.
    l := ctxlogger_logrus.Extract(ctx)
    l.Info("some ping")
    l.Info("another ping")
    return &pb_testproto.PingResponse{Value: ping.Value}, nil
}
return x
```

</details>

#### Example:

<details>
<summary>Click to expand code.</summary>

```go
// Logrus entry is used, allowing pre-definition of certain fields by the user.
logrusEntry := logrus.NewEntry(logrusLogger)

// Create a server, make sure we put the grpc_ctxtags context before everything else.
server := grpc.NewServer(
    grpc_middleware.WithUnaryServerChain(
        grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
        ctxlogger_logrus.UnaryServerInterceptor(logrusEntry),
    ),
    grpc_middleware.WithStreamServerChain(
        grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
        ctxlogger_logrus.StreamServerInterceptor(logrusEntry),
    ),
)
return server
```

</details>

## <a name="pkg-imports">Imported Packages</a>

- [github.com/grpc-ecosystem/go-grpc-middleware](./../../..)
- [github.com/grpc-ecosystem/go-grpc-middleware/tags](./../..)
- [github.com/sirupsen/logrus](https://godoc.org/github.com/sirupsen/logrus)
- [golang.org/x/net/context](https://godoc.org/golang.org/x/net/context)
- [google.golang.org/grpc](https://godoc.org/google.golang.org/grpc)

## <a name="pkg-index">Index</a>
* [func AddFields(ctx context.Context, fields logrus.Fields)](#AddFields)
* [func Extract(ctx context.Context) \*logrus.Entry](#Extract)
* [func StreamServerInterceptor(entry \*logrus.Entry) grpc.StreamServerInterceptor](#StreamServerInterceptor)
* [func ToContext(ctx context.Context, entry \*logrus.Entry) context.Context](#ToContext)
* [func UnaryServerInterceptor(entry \*logrus.Entry) grpc.UnaryServerInterceptor](#UnaryServerInterceptor)

#### <a name="pkg-examples">Examples</a>
* [Package (HandlerUsageUnaryPing)](#example__handlerUsageUnaryPing)
* [Package (Initialization)](#example__initialization)

#### <a name="pkg-files">Package files</a>
[context.go](./context.go) [doc.go](./doc.go) [noop.go](./noop.go) [server_interceptors.go](./server_interceptors.go) 

## <a name="AddFields">func</a> [AddFields](./context.go#L21)
``` go
func AddFields(ctx context.Context, fields logrus.Fields)
```
AddFields adds logrus fields to the logger.

## <a name="Extract">func</a> [Extract](./context.go#L35)
``` go
func Extract(ctx context.Context) *logrus.Entry
```
Extract takes the call-scoped logrus.Entry from ctxlogger_logrus middleware.

If the ctxlogger_logrus middleware wasn't used, a no-op `logrus.Entry` is returned. This makes it safe to
use regardless.

## <a name="StreamServerInterceptor">func</a> [StreamServerInterceptor](./server_interceptors.go#L20)
``` go
func StreamServerInterceptor(entry *logrus.Entry) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that adds logrus.Entry to the context.

## <a name="ToContext">func</a> [ToContext](./context.go#L57)
``` go
func ToContext(ctx context.Context, entry *logrus.Entry) context.Context
```

## <a name="UnaryServerInterceptor">func</a> [UnaryServerInterceptor](./server_interceptors.go#L11)
``` go
func UnaryServerInterceptor(entry *logrus.Entry) grpc.UnaryServerInterceptor
```
PayloadUnaryServerInterceptor returns a new unary server interceptors that adds logrus.Entry to the context.

- - -
Generated by [godoc2ghmd](https://github.com/GandalfUK/godoc2ghmd)