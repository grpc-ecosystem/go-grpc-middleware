/*
Package ctxzerolog is a ctxlogger that is backed by zerolog.

It accepts a user-configured `zerolog.Context` that will be used for logging. The same Context will
be populated into the `context.Context` passed into gRPC handler code.

You can use `ctxzerolog.Extract` to log into a request-scoped Context instance in your handler code.

As `ctxzerolog.Extract` will iterate on all tags from `grpc_ctxtags` it is therefore expensive so it
is advised that you extract once at the start of the function from the context and reuse it for the
remainder of the function (see examples).

Please see examples and tests for examples of use.
*/
package ctxzerolog
