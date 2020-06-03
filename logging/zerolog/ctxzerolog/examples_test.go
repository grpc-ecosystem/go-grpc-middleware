package ctxzerolog_test

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zerolog/ctxzerolog"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	ctx := context.Background()
	// setting tags will be added to the logger as log fields
	grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	// Extract a single request-scoped zerolog.Context and log messages.
	l := ctxzerolog.Extract(ctx).Logger()
	l.Info().Msg("some ping")
	l.Info().Msg("another ping")
}
