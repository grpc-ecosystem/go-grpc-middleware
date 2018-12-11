// Copyright 2018 Zheng Dayu. All Rights Reserved.
// See LICENSE for licensing terms.

/*
`grpc_ratelimit` a generic server-side ratelimit middleware for gRPC.

Server Side Ratelimit Middleware

It allows to use your own rate limiter (e.g. token bucket, leaky bucket, etc.) to do grpc rate limit.

`ratelimit/tokenbucket`provides an implementation based on  token bucket `github.com/juju/ratelimit`.

Please see examples for simple examples of use.
*/
package grpc_ratelimit
