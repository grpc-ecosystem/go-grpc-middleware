/*
`ctxlogger_logrus` is a ctxlogger that is backed by logrus

It accepts a user-configured `logrus.Logger` that will be used for logging. The same `logrus.Logger` will
be populated into the `context.Context` passed into gRPC handler code.

On calling `StreamServerInterceptor` or `UnaryServerInterceptor` this ctxlogger middleware will add the Tags from
the ctx so that it will be present on subsequent use of the `ctxlogger_logrus` logger.

You can use `ctxlogger_zap.Extract` to log into a request-scoped `zap.Logger` instance in your handler code.

As `ctxlogger_zap.Extract` will iterate all tags on from `grpc_ctxtags` it is therefore expensive so it is advised that you
extract once at the start of the function from the context and reuse it for the remainder of the function (see examples).

Please see examples and tests for examples of use.
*/
package ctxlogger_logrus
