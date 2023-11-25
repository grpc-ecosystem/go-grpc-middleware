// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip_test

import (
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"google.golang.org/grpc"
)

// Simple example of a unary server initialization code.
func ExampleUnaryServerInterceptor() {
	// Define list of trusted peers from which we accept forwarded-for and
	// real-ip headers.
	trustedPeers := []net.IPNet{
		{IP: net.IPv4(127, 0, 0, 1), Mask: net.IPv4Mask(255, 255, 255, 255)},
	}
	// Define headers to look for in the incoming request.
	headers := []string{"x-forwarded-for", "x-real-ip"}
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
	trustedPeers := []net.IPNet{
		{IP: net.IPv4(127, 0, 0, 1), Mask: net.IPv4Mask(255, 255, 255, 255)},
	}
	// Define headers to look for in the incoming request.
	headers := []string{"x-forwarded-for", "x-real-ip"}
	_ = grpc.NewServer(
		grpc.ChainStreamInterceptor(
			realip.StreamServerInterceptor(trustedPeers, headers),
		),
	)
}
