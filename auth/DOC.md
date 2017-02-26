# grpc_auth
--
    import "github.com/mwitkow/go-grpc-middleware/auth"

`grpc_auth` a generic server-side auth middleware for gRPC.


### Server Side Auth Middleware

It allows for easy assertion of `:authorization` headers in gRPC calls, be it
HTTP Basic auth, or OAuth2 Bearer tokens.

The middleware takes a user-customizable `AuthFunc`, which can be customized to
verify and extract auth information from the request. The extracted information
can be put in the `context.Context` of handlers downstream for retrieval.

It also allows for per-service implementation overrides of `AuthFunc`. See
`ServiceAuthFuncOverride`.

Please see examples for simple examples of use.

## Usage

#### func  AuthFromMD

```go
func AuthFromMD(ctx context.Context, expectedScheme string) (string, error)
```
AuthFromMD is a helper function for extracting the :authorization header from
the gRPC metadata of the request.

It expects the `:authorization` header to be of a certain scheme (e.g. `basic`,
`bearer`), in a case-insensitive format (see rfc2617, sec 1.2). If no such
authorization is found, or the token is of wrong scheme, an error with gRPC
status `Unauthenticated` is returned.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(authFunc AuthFunc) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new unary server interceptors that performs
per-request auth.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(authFunc AuthFunc) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptors that performs
per-request auth.

#### type AuthFunc

```go
type AuthFunc func(ctx context.Context) (context.Context, error)
```

AuthFunc is the pluggable function that performs authentication.

The passed in `Context` will contain the gRPC metadata.MD object (for
header-based authentication) and the peer.Peer information that can contain
transport-based credentials (e.g. `credentials.AuthInfo`).

The returned context will be propagated to handlers, allowing user changes to
`Context`. However, please make sure that the `Context` returned is a child
`Context` of the one passed in.

If error is returned, its `grpc.Code()` will be returned to the user as well as
the verbatim message. Please make sure you use `codes.Unauthenticated` (lacking
auth) and `codes.PermissionDenied` (authed, but lacking perms) appropriately.

#### type ServiceAuthFuncOverride

```go
type ServiceAuthFuncOverride interface {
	AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error)
}
```

ServiceAuthFuncOverride allows a given gRPC service implementation to override
the global `AuthFunc`.

If a service implements the AuthFuncOverride method, it takes precedence over
the `AuthFunc` method, and will be called instead of AuthFunc for all method
invocations within that service.
