// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`grpc_ctxtags` adds a Tag object to the context that can be used by other middleware to add context about a request.

Request Context Tags

Tags describe information about the request, and can be set and used by other middleware. Tags are used for logging
and tracing of requests.

Tags fields are typed, and shallow and should follow the OpenTracing semantics convention:
https://github.com/opentracing/specification/blob/master/semantic_conventions.md

*/

package grpc_ctxtags
