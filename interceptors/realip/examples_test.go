// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip_test

import (
	"net/netip"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"google.golang.org/grpc"
)

// Simple example of a unary server initialization code.
func ExampleUnaryServerInterceptorOpts() {
	// Define list of trusted peers from which we accept forwarded-for and
	// real-ip headers.
	trustedPeers := []netip.Prefix{
		netip.MustParsePrefix("127.0.0.1/32"),
	}
	// Define headers to look for in the incoming request.
	headers := []string{realip.XForwardedFor, realip.XRealIp}
	// Consider that there is one proxy in front,
	// so the real client ip will be rightmost - 1 in the csv list of X-Forwarded-For
	// Optionally you can specify TrustedProxies
	trustedProxyCnt := uint(1)
	opts := realip.Opts{TrustedPeers: trustedPeers, Headers: headers, TrustedProxiesCount: trustedProxyCnt}
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			realip.UnaryServerInterceptorOpts(opts),
		),
	)
}

// Simple example of a streaming server initialization code.
func ExampleStreamServerInterceptorOpts() {
	// Define list of trusted peers from which we accept forwarded-for and
	// real-ip headers.
	trustedPeers := []netip.Prefix{
		netip.MustParsePrefix("127.0.0.1/32"),
	}
	// Define headers to look for in the incoming request.
	headers := []string{realip.XForwardedFor, realip.XRealIp}
	// Consider that there is one proxy in front,
	// so the real client ip will be rightmost - 1 in the csv list of X-Forwarded-For
	// Optionally you can specify TrustedProxies
	trustedProxyCnt := uint(1)
	opts := realip.Opts{TrustedPeers: trustedPeers, Headers: headers, TrustedProxiesCount: trustedProxyCnt}
	_ = grpc.NewServer(
		grpc.ChainStreamInterceptor(
			realip.StreamServerInterceptorOpts(opts),
		),
	)
}
