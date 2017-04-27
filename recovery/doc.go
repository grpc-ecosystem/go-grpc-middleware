// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`grpc_recovery` conversion of panics into gRPC errors

Server Side Recovery Middleware

By default a panic will be converted into a gRPC error with `code.Internal`.

Handling can be customised by providing an alternate recovery function.

Please see examples for simple examples of use.
*/
package grpc_recovery
