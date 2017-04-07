# grpc_ctxtags
--
    import "github.com/grpc-ecosystem/go-grpc-middleware/tags"


## Usage

#### func  CodeGenRequestFieldExtractor

```go
func CodeGenRequestFieldExtractor(fullMethod string, req interface{}) map[string]interface{}
```
CodeGenRequestFieldExtractor is a function that relies on code-generated
functions that export log fields from requests. These are usually coming from a
protoc-plugin that generates additional information based on custom field
options.

#### func  StreamServerInterceptor

```go
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor
```
StreamServerInterceptor returns a new streaming server interceptor that sets the
values for request tags.

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptors that sets the
values for request tags.

#### type Option

```go
type Option func(*options)
```


#### func  WithFieldExtractor

```go
func WithFieldExtractor(f RequestFieldExtractorFunc) Option
```
WithFieldExtractor customizes the function for extracting log fields from
protobuf messages.

#### type RequestFieldExtractorFunc

```go
type RequestFieldExtractorFunc func(fullMethod string, req interface{}) map[string]interface{}
```

RequestFieldExtractorFunc is a user-provided function that extracts field
information from a gRPC request. It is called from tags middleware on arrival of
unary request or a server-stream request. Keys and values will be added to the
context tags of the request. If there are no fields, you should return a nil.

#### func  TagBasedRequestFieldExtractor

```go
func TagBasedRequestFieldExtractor(tagName string) RequestFieldExtractorFunc
```
TagedRequestFiledExtractor is a function that relies on Go struct tags to export
log fields from requests. These are usualy coming from a protoc-plugin, such as
Gogo protobuf.

    message Metadata {
       repeated string tags = 1 [ (gogoproto.moretags) = "log_field:\"meta_tags\"" ];
    }

The tagName is configurable using the tagName variable. Here it would be
"log_field".

#### type Tags

```go
type Tags struct {
}
```

Tags is the struct used for storing request tags between Context calls. This
object is *not* thread safe, and should be handled only in the context of the
request.

#### func  Extract

```go
func Extract(ctx context.Context) *Tags
```
Extracts returns a pre-existing Tags object in the Context. If the context
wasn't set in a tag interceptor, a no-op Tag storage is returned that will *not*
be propagated in context.

#### func (*Tags) Has

```go
func (t *Tags) Has(key string) bool
```
Has checks if the given key exists.

#### func (*Tags) Set

```go
func (t *Tags) Set(key string, value interface{}) *Tags
```
Set sets the given key in the metadata tags.

#### func (*Tags) Values

```go
func (t *Tags) Values() map[string]interface{}
```
Values returns a map of key to values. Do not modify the underlying map, please
use Set instead.
