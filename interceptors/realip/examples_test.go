// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip_test

import (
	"net/netip"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"google.golang.org/grpc"
)

// Simple example of a unary server initialization code.
func ExampleUnaryServerInterceptor() {
	// Define list of trusted peers from which we accept forwarded-for and
	// real-ip headers.
	trustedPeers := []netip.Prefix{
		netip.MustParsePrefix("127.0.0.1/32"),
	}
	// Define headers to look for in the incoming request.
	headers := []string{realip.X_FORWARDED_FOR, realip.X_REAL_IP}
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			realip.UnaryServerInterceptor(trustedPeers, headers),
		),
	)
}

// Simple example of a streaming server initialization code.
func ExampleStreamServerInterceptor() {
	// Define list of trusted peers from which we accept forwarded-for and
	// real-ip headers.
	trustedPeers := []netip.Prefix{
		netip.MustParsePrefix("127.0.0.1/32"),
	}
	// Define headers to look for in the incoming request.
	headers := []string{realip.X_FORWARDED_FOR, realip.X_REAL_IP}
	_ = grpc.NewServer(
		grpc.ChainStreamInterceptor(
			realip.StreamServerInterceptor(trustedPeers, headers),
		),
	)
}
