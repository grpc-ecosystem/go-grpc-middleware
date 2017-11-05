// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_ctxtags

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// RequestMetadataExtractorFunc is a user-provided function that extracts field information from a gRPC request.
// It is called from tags middleware on arrival of unary request or a server-stream request.
// Keys and values will be added to the context tags of the request.
type RequestMetadataExtractorFunc func(ctx context.Context, req interface{})

func TagBasedRequestMetadataExtractor(prefix string, fields ...string) RequestMetadataExtractorFunc {
	return func(ctx context.Context, _ interface{}) {
		if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
			tags := Extract(ctx)
			for _, field := range fields {
				if values, present := ctxMd[field]; present {
					tags = tags.Set(prefix+field, strings.Join(values, ","))
				}
			}
		}
	}
}
