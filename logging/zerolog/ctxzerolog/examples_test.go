package ctxzerolog_test

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/irridia/go-grpc-middleware/logging/zerolog/ctxzerolog"
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
