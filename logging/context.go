package grpc_logging

import "context"

type grpcLoggerMarker struct {
}

var (
	// InternalContextMarker is the Context value marker used by *all* logging middleware.
	// The logging middleware object must interf
	InternalContextMarker = &grpcLoggerMarker{}

	noOp = &noOpMetadata{}
)

// Metadata is a common interface for interacting with the request-scope of a logger provided by any middleware.
type Metadata interface {
	AddFieldsFromMiddleware(keys []string, values []interface{})
}

// ExtractMetadata allows other middleware to access the metadata (e.g. request-scope fields) of any logging middleware.
func ExtractMetadata(ctx context.Context) Metadata {
	md, ok := ctx.Value(InternalContextMarker).(Metadata)
	if !ok {
		return noOp
	}
	return md
}

type noOpMetadata struct {
}

func (*noOpMetadata) AddFieldsFromMiddleware(keys []string, values []interface{}) {
}
