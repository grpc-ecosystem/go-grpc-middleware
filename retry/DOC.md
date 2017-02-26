# grpc_retry
--
    import "github.com/mwitkow/go-grpc-middleware/retry"

`grpc_retry` provides client-side request retry logic for gRPC.


### Client-Side Request Retry Interceptor

It allows for automatic retry, inside the generated gRPC code of requests based
on the gRPC status of the reply. It supports unary (1:1), and server stream
(1:n) requests.

By default the interceptors *are disabled*, preventing accidental use of
retries. You can easily override the number of retries (setting them to more
than 0) with a `grpc.ClientOption`, e.g.:

    myclient.Ping(ctx, goodPing, grpc_retry.WithMax(5))

Other default options are: retry on `ResourceExhausted` and `Unavailable` gRPC
codes, use a 50ms linear backoff with 10% jitter.

Please see examples for more advanced use.

## Usage

```go
const (
	AttemptMetadataKey = "x-retry-attempty"
)
```

```go
var (
	// DefaultRetriableCodes is a set of well known types gRPC codes that should be retri-able.
	//
	// `ResourceExhausted` means that the user quota, e.g. per-RPC limits, have been reached.
	// `Unavailable` means that system is currently unavailable and the client should retry again.
	DefaultRetriableCodes = []codes.Code{codes.ResourceExhausted, codes.Unavailable}
)
```

#### func  StreamClientInterceptor

```go
func StreamClientInterceptor(optFuncs ...CallOption) grpc.StreamClientInterceptor
```
StreamClientInterceptor returns a new retrying stream client interceptor for
server side streaming calls.

The default configuration of the interceptor is to not retry *at all*. This
behaviour can be changed through options (e.g. WithMax) on creation of the
interceptor or on call (through grpc.CallOptions).

Retry logic is available *only for ServerStreams*, i.e. 1:n streams, as the
internal logic needs to buffer the messages sent by the client. If retry is
enabled on any other streams (ClientStreams, BidiStreams), the retry interceptor
will fail the call.

#### func  UnaryClientInterceptor

```go
func UnaryClientInterceptor(optFuncs ...CallOption) grpc.UnaryClientInterceptor
```
UnaryClientInterceptor returns a new retrying unary client interceptor.

The default configuration of the interceptor is to not retry *at all*. This
behaviour can be changed through options (e.g. WithMax) on creation of the
interceptor or on call (through grpc.CallOptions).

#### type BackoffFunc

```go
type BackoffFunc func(attempt uint) time.Duration
```

BackoffFunc denotes a family of functions that controll the backoff duration
between call retries.

They are called with an identifier of the attempt, and should return a time the
system client should hold off for. If the time returned is longer than the
`context.Context.Deadline` of the request the deadline of the request takes
precedence and the wait will be interrupted before proceeding with the next
iteration.

#### func  BackoffLinear

```go
func BackoffLinear(waitBetween time.Duration) BackoffFunc
```
BackoffLinear is very simple: it waits for a fixed period of time between calls.

#### func  BackoffLinearWithJitter

```go
func BackoffLinearWithJitter(waitBetween time.Duration, jitterFraction float64) BackoffFunc
```
BackoffLinearWithJitter waits a set period of time, allowing for jitter
(fractional adjustment).

For example waitBetween=1s and jitter=0.10 can generate waits between 900ms and
1100ms.

#### type CallOption

```go
type CallOption struct {
	grpc.CallOption // anonymously implement it, without knowing the private fields.
}
```

callOption is a grpc.CallOption that is local to grpc_retry.

#### func  Disable

```go
func Disable() CallOption
```
Disable disables the retry behaviour on this call, or this interceptor.

Its semantically the same to `WithMax`

#### func  WithBackoff

```go
func WithBackoff(bf BackoffFunc) CallOption
```
WithBackoff sets the `BackoffFunc `used to control time between retries.

#### func  WithCodes

```go
func WithCodes(retryCodes ...codes.Code) CallOption
```
WithCodes sets which codes should be retried.

Please *use with care*, as you may be retrying non-idempotend calls.

You cannot automatically retry on Cancelled and Deadline, please use
`WithPerRetryTimeout` for these.

#### func  WithMax

```go
func WithMax(maxRetries uint) CallOption
```
WithMax sets the maximum number of retries on this call, or this interceptor.

#### func  WithPerRetryTimeout

```go
func WithPerRetryTimeout(timeout time.Duration) CallOption
```
WithPerRetryTimeout sets the RPC timeout per call (including initial call) on
this call, or this interceptor.

The context.Deadline of the call takes precedence and sets the maximum time the
whole invocation will take, but WithPerCallTimeout can be used to limit the RPC
time per each call.

For example, with context.Deadline = now + 10s, and WithPerCallTimeout(3 *
time.Seconds), each of the retry calls (including the initial one) will have a
deadline of now + 3s.

A value of 0 disables the timeout overrides completely and returns to each retry
call using the parent `context.Deadline`.
