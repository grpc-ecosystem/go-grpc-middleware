module github.com/grpc-ecosystem/go-grpc-middleware/grpctesting/external/v2

go 1.14

// We maintain grpctesting/external as separate Go module, so we don't pull this dependency when someone depends on middleware core.
require (
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-20200501113911-9a95f0fdbfea
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.30.0
)
