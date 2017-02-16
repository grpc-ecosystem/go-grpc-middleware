// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_auth

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	headerAuthorize = "authorization"
)

// AuthFromMD is a helper function for extracting the :authorization header from the gRPC metadata of the request.
//
// It expects the `:authorization` header to be of a certain scheme (e.g. `basic`, `bearer`), in a
// case-insensitive format (see rfc2617, sec 1.2). If no such authorization is found, or the token
// is of wrong scheme, an error with gRPC status `Unauthenticated` is returned.
func AuthFromMD(ctx context.Context, expectedScheme string) (string, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", grpc.Errorf(codes.Internal, "gRPC context lacks metadata")
	}
	slice, ok := md[headerAuthorize]
	if !ok {
		return "", grpc.Errorf(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)
	}
	if len(slice) > 1 {
		return "", grpc.Errorf(codes.Unauthenticated, "Request contains multiple authorization headers")
	}
	splits := strings.SplitN(slice[0], " ", 2)
	if len(splits) < 2 {
		return "", grpc.Errorf(codes.Unauthenticated, "Bad authorization string")
	}
	if strings.ToLower(splits[0]) != strings.ToLower(expectedScheme) {
		return "", grpc.Errorf(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)
	}
	return splits[1], nil
}
