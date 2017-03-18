# grpc_logging
--
    import "github.com/mwitkow/go-grpc-middleware/logging"

grpc_logging is a "parent" package for gRPC logging middlewares


### General functionality of all middleware

All logging middleware have an `Extract(ctx)` function that provides a
request-scoped logger with gRPC-related fields (service and method names).
Additionally, in case a `WithFieldExtractor` is used, the logger will have
fields extracted from the content of the inbound request (unary and server-side
stream).

All logging middleware will emit a final log statement. It is based on the error
returned by the handler function, the gRPC status code (controlled via `WithCodes`), an error (if any) and it
will emit at a level controlled via `WithLevels`.


### This parent package

This particular package is intended for use by other middleware, logging or
otherwise. It contains interfaces that other logging middlewares *should*
implement. This allows code to be shared between different implementations.

The `RequestLogFieldExtractorFunc` signature allows users to customize the
extraction of request fields to be used as log fields in middlewares. There are
two implementations: one (default) that relies on optional code-generated
`ExtractLogFields()` methods on protobuf structs, and another that uses tagging.


### Implementations

There are two implementations at the moment: logrus and zap

See relevant packages below.

## Usage

```go
var (
	// InternalContextMarker is the Context value marker used by *all* logging middleware.
	// The logging middleware object must interf
	InternalContextMarker = &grpcLoggerMarker{}
)
```

#### func  CodeGenRequestLogFieldExtractor

```go
func CodeGenRequestLogFieldExtractor(fullMethod string, req interface{}) (keys []string, values []interface{})
```
CodeGenRequestLogFieldExtractor is a function that relies on code-generated
functions that export log fields from requests. These are usually coming from a
protoc-plugin that generates additional information based on custom field
options.

#### func  TagedRequestFiledExtractor

```go
func TagedRequestFiledExtractor(fullMethod string, req interface{}) (keys []string, values []interface{})
```
TagedRequestFiledExtractor is a function that relies on Go struct tags to export
log fields from requests. These are usualy coming from a protoc-plugin, such as
Gogo protobuf.

    message Metadata {
       repeated string tags = 1 [ (gogoproto.moretags) = "log_field:\"meta_tags\"" ];
    }

It requires the tag to be `log_field` and is recursively executed through all
non-repeated structs.

#### type Metadata

```go
type Metadata interface {
	AddFieldsFromMiddleware(keys []string, values []interface{})
}
```

Metadata is a common interface for interacting with the request-scope of a
logger provided by any middleware.

#### func  ExtractMetadata

```go
func ExtractMetadata(ctx context.Context) Metadata
```
ExtractMetadata allows other middleware to access the metadata (e.g.
request-scope fields) of any logging middleware.

#### type RequestLogFieldExtractorFunc

```go
type RequestLogFieldExtractorFunc func(fullMethod string, req interface{}) (keys []string, values []interface{})
```

RequestLogFieldExtractorFunc is a user-provided function that extracts field
information from a gRPC request. It is called from every logging middleware on
arrival of unary request or a server-stream request. Keys and values will be
added to the logging request context.
