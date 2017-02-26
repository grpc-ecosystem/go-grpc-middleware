# grpc_logging
--
    import "github.com/mwitkow/go-grpc-middleware/logging"

gRPC middleware logging.

`grpc_logging` is a "mother" package for other specific gRPC logging middleware.

General functionality across all logging middleware:

    * Extract(ctx) function that provides a request-scoped logger with pre-defined fields
    * log statement on completion of handling with customizeable log levels, gRPC status code and error message logging
    * automatic request field to log field extraction, either through code-generated data or field annotations

### Concrete logging middleware for use in user-code handlers is available in
### subpackages

    * zap
    * logrus

### The functions and methods in this package should only be consumed by gRPC
### logging middleware and other middlewares that want to add metadata to the
logging context of the request.

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
