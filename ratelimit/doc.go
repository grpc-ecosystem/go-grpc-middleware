// See LICENSE for licensing terms.

/*
`ratelimit` a generic server-side ratelimit middleware for gRPC.

Server Side Ratelimit Middleware

It allows to do grpc rate limit by your own rate limiter (e.g. token bucket, leaky bucket, etc.)

`ratelimit/tokenbucket`provides an implementation based on token bucket `github.com/juju/ratelimit`.

Please see examples for simple examples of use.
*/
package ratelimit
