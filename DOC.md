# grpc_middleware
--
    import "github.com/grpc-ecosystem/go-grpc-middleware"

`grpc_middleware` is a collection of gRPC middleware packages: interceptors,
helpers and tools.


### Middleware

gRPC is a fantastic RPC middleware, which sees a lot of adoption in the Golang
world. However, the upstream gRPC codebase is relatively bare bones.

This package, and most of its child packages provides commonly needed middleware
for gRPC: client-side interceptors for retires, server-side interceptors for
input validation and auth, functions for chaining said interceptors, metadata
convenience methods and more.


### Chaining

By default, gRPC doesn't allow one to have more than one interceptor either on
the client nor on the server side. `grpc_middleware` provides convenient
### chaining methods

Simple way of turning a multiple interceptors into a single interceptor. Here's
an example for server chaining:

    myServer := grpc.NewServer(
        grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(loggingStream, monitoringStream, authStream)),
        grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(loggingUnary, monitoringUnary, authUnary),
    )

These interceptors will be executed from left to right: logging, monitoring and
auth.

Here's an example for client side chaining:

    clientConn, err = grpc.Dial(
        address,
            grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(monitoringClientUnary, retryUnary)),
            grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(monitoringClientStream, retryStream)),
    )
    client = pb_testproto.NewTestServiceClient(clientConn)
    resp, err := client.PingEmpty(s.ctx, &myservice.Request{Msg: "hello"})

These interceptors will be executed from left to right: monitoring and then
retry logic.


### Writing Your Own

Implementing your own interceptor is pretty trivial: there are interfaces for
that. But the interesting bit exposing common data to handlers (and other
middleware), similarly to HTTP Middleware design. For example, you may want to
### pass the identity of the caller from the auth interceptor all the way to the
handling function.

For example, a client side interceptor example for auth looks like:

    func FakeAuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
       newCtx := context.WithValue(ctx, "user_id", "john@example.com")
       return handler(newCtx, req)
    }

Unfortunately, it's not as easy for streaming RPCs. These have the
`context.Context` embedded within the `grpc.ServerStream` object. To pass values
through context, a wrapper (`WrappedServerStream`) is needed. For example:

    func FakeAuthStreamingInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
       newStream := grpc_middleware.WrapServerStream(stream)
       newStream.WrappedContext = context.WithValue(ctx, "user_id", "john@example.com")
       return handler(srv, stream)
    }

## Usage

#### func  ChainStreamClient

```go
func ChainStreamClient(interceptors ...grpc.StreamClientInterceptor) grpc.StreamClientInterceptor
```
ChainStreamClient creates a single interceptor out of a chain of many
interceptors.

Execution is done in left-to-right order, including passing of context. For
example ChainStreamClient(one, two, three) will execute one before two before
three.

#### func  ChainStreamServer

```go
func ChainStreamServer(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor
```
ChainStreamServer creates a single interceptor out of a chain of many
interceptors.

Execution is done in left-to-right order, including passing of context. For
example ChainUnaryServer(one, two, three) will execute one before two before
three. If you want to pass context between interceptors, use WrapServerStream.

#### func  ChainUnaryClient

```go
func ChainUnaryClient(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor
```
ChainUnaryClient creates a single interceptor out of a chain of many
interceptors.

Execution is done in left-to-right order, including passing of context. For
example ChainUnaryClient(one, two, three) will execute one before two before
three.

#### func  ChainUnaryServer

```go
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor
```
ChainUnaryServer creates a single interceptor out of a chain of many
interceptors.

Execution is done in left-to-right order, including passing of context. For
example ChainUnaryServer(one, two, three) will execute one before two before
three, and three will see context changes of one and two.

#### func  WithStreamServerChain

```go
func WithStreamServerChain(interceptors ...grpc.StreamServerInterceptor) grpc.ServerOption
```
WithStreamServerChain is a grpc.Server config option that accepts multiple
stream interceptors. Basically syntactic sugar.

#### func  WithUnaryServerChain

```go
func WithUnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption
```
Chain creates a single interceptor out of a chain of many interceptors.

WithUnaryServerChain is a grpc.Server config option that accepts multiple unary
interceptors. Basically syntactic sugar.

#### type WrappedServerStream

```go
type WrappedServerStream struct {
	grpc.ServerStream
	// WrappedContext is the wrapper's own Context. You can assign it.
	WrappedContext context.Context
}
```

WrappedServerStream is a thin wrapper around grpc.ServerStream that allows
modifying context.

#### func  WrapServerStream

```go
func WrapServerStream(stream grpc.ServerStream) *WrappedServerStream
```
WrapServerStream returns a ServerStream that has the ability to overwrite
context.

#### func (*WrappedServerStream) Context

```go
func (w *WrappedServerStream) Context() context.Context
```
Context returns the wrapper's WrappedContext, overwriting the nested
grpc.ServerStream.Context()
