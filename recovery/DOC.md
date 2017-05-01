# grpc_recovery
--
    import "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

`grpc_recovery` conversion of panics into gRPC errors


### Server Side Recovery Middleware

By default a panic will be converted into a gRPC error with `code.Internal`.

Handling can be customised by providing an alternate recovery function.

Please see examples for simple examples of use.

## Usage

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor for panic
recovery.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptor for panic
recovery.

#### type Option

```go
type Option func(*options)
```


#### func  WithRecoveryHandler

```go
func WithRecoveryHandler(f RecoveryHandlerFunc) Option
```
WithRecoveryHandler customizes the function for recovering from a panic.

#### type RecoveryHandlerFunc

```go
type RecoveryHandlerFunc func(p interface{}) (err error)
```

RecoveryHandlerFunc is a function that recovers from the panic `p` by returning
an `error`.
