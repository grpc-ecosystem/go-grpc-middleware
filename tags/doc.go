// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`grpc_ctxtags` adds a Tag object to the context that can be used by other middleware to add context about a request.

Request Context Tags

Tags describe information about the request, and can be set and used by other middleware, or handlers. Tags are used
for logging and tracing of requests. Tags are populated both upwards, *and* downwards in the interceptor-handler stack.

You can automatically extract tags (in `grpc.request.<field_name>`) from request payloads (in unary and server-streaming)
by passing in the `WithFieldExtractor` option.

If a user doesn't use the interceptors that initialize the `Tags` object, all operations following from an `Extract(ctx)`
will be no-ops. This is to ensure that code doesn't panic if the interceptors weren't used.

Tags fields are typed, and shallow and should follow the OpenTracing semantics convention:
https://github.com/opentracing/specification/blob/master/semantic_conventions.md
*/
package grpc_ctxtags
