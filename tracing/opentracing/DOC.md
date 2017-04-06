# grpc_opentracing
--
    import "github.com/mwitkow/go-grpc-middleware/tracing/opentracing"


## Usage

#### func  StreamClientInterceptor

```go
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor
```
StreamClientInterceptor returns a new streaming server interceptor for
OpenTracing.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor for
OpenTracing.

#### func  UnaryClientInterceptor

```go
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor
```
UnaryClientInterceptor returns a new unary server interceptor for OpenTracing.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptor for OpenTracing.

#### type FilterOutFunc

```go
type FilterOutFunc func(ctx context.Context, fullMethodName string) bool
```

FilterOutFunc allows users to provide a function that filters out certain
methods from being traced.

If it returns false, the given request will not be traced.

#### type Option

```go
type Option func(*options)
```


#### func  WithFilterOutFunc

```go
func WithFilterOutFunc(f FilterOutFunc) Option
```
WithFilterOutFunc customizes the function used for deciding

#### func  WithTracer

```go
func WithTracer(tracer opentracing.Tracer) Option
```
WithTracer sets a custom tracer to be used for this middleware, otherwise the
opentracing.GlobalTracer is used.
