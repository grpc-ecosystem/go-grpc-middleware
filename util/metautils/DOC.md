# metautils
--
    import "github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"


## Usage

#### func  GetSingle

```go
func GetSingle(ctx context.Context, keyName string) (string, bool)
```
GetSingle extracts a single-value metadata key from Context. First return is the
value of the key, followed by a bool indicator. The bool indicator being false
means the string should be discarded. It can be false if the context has no
metadata at all, the key in metadata doesn't exist or there are multiple values.
Deprecated, use NiceMD.Get.

#### func  SetSingle

```go
func SetSingle(ctx context.Context, keyName string, keyValue string) context.Context
```
SetSingle sets or overrides a metadata key to be single value in the Context. It
returns a new context.Context object that contains a *copy* of the metadata
inside the given context. Deprecated, use NiceMD.Set.

#### type NiceMD

```go
type NiceMD metadata.MD
```

NiceMD is a convenience wrapper definiting extra functions on the metadata.

#### func  ExtractIncoming

```go
func ExtractIncoming(ctx context.Context) NiceMD
```
ExtractIncoming extracts an inbound metadata from the server-side context.

This function always returns a NiceMD wrapper of the metadata.MD, in case the
context doesn't have metadata it returns a new empty NiceMD.

#### func  ExtractOutgoing

```go
func ExtractOutgoing(ctx context.Context) NiceMD
```
ExtractOutgoing extracts an outbound metadata from the client-side context.

This function always returns a NiceMD wrapper of the metadata.MD, in case the
context doesn't have metadata it returns a new empty NiceMD.

#### func (NiceMD) Add

```go
func (m NiceMD) Add(key string, value string) NiceMD
```
Add retrieves a single value from the metadata.

It works analogously to http.Header.Add, as it appends to any existing values
associated with key.

The function is binary-key safe.

#### func (NiceMD) Clone

```go
func (m NiceMD) Clone(copiedKeys ...string) NiceMD
```
Clone performs a *deep* copy of the metadata.MD.

You can specify the lower-case copiedKeys to only copy certain whitelisted keys.
If no keys are explicitly whitelisted all keys get copied.

#### func (NiceMD) Del

```go
func (m NiceMD) Del(key string) NiceMD
```

#### func (NiceMD) Get

```go
func (m NiceMD) Get(key string) string
```
Get retrieves a single value from the metadata.

It works analogously to http.Header.Get, returning the first value if there are
many set. If the value is not set, an empty string is returned.

The function is binary-key safe.

#### func (NiceMD) Set

```go
func (m NiceMD) Set(key string, value string) NiceMD
```
Set sets the given value in a metadata.

It works analogously to http.Header.Set, overwriting all previous metadata
values.

The function is binary-key safe.

#### func (NiceMD) ToIncoming

```go
func (m NiceMD) ToIncoming(ctx context.Context) context.Context
```
ToIncoming sets the given NiceMD as a server-side context for dispatching.

This is mostly useful in ServerInterceptors..

#### func (NiceMD) ToOutgoing

```go
func (m NiceMD) ToOutgoing(ctx context.Context) context.Context
```
ToOutgoing sets the given NiceMD as a client-side context for dispatching.
