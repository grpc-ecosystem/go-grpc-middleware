// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip

import (
	"context"
	"net"
	"net/netip"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// XRealIp, XForwardedFor and TrueClientIp are header keys
// used to extract the real client IP from the request. They represent common
// conventions for identifying the originating IP address of a client connecting
// through proxies or load balancers.
const (
	XRealIp       = "X-Real-IP"
	XForwardedFor = "X-Forwarded-For"
	TrueClientIp  = "True-Client-IP"
)

var noIP = netip.Addr{}

type realipKey struct{}

// FromContext extracts the real client IP from the context.
// It returns the IP and a boolean indicating if it was present.
func FromContext(ctx context.Context) (netip.Addr, bool) {
	ip, ok := ctx.Value(realipKey{}).(netip.Addr)
	return ip, ok
}

func remotePeer(ctx context.Context) net.Addr {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	return pr.Addr
}

func ipInNets(ip netip.Addr, nets []netip.Prefix) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func getHeader(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	if md[strings.ToLower(key)] == nil {
		return ""
	}

	return md[strings.ToLower(key)][0]
}

func ipFromHeaders(ctx context.Context, headers []string) netip.Addr {
	for _, header := range headers {
		a := strings.Split(getHeader(ctx, header), ",")
		h := strings.TrimSpace(a[len(a)-1])
		ip, err := netip.ParseAddr(h)
		if err == nil {
			return ip
		}
	}
	return noIP
}

func getRemoteIP(ctx context.Context, trustedPeers []netip.Prefix, headers []string) netip.Addr {
	pr := remotePeer(ctx)
	if pr == nil {
		return noIP
	}

	addrPort, err := netip.ParseAddrPort(pr.String())
	if err != nil {
		return noIP
	}
	ip := addrPort.Addr()

	if len(trustedPeers) == 0 || !ipInNets(ip, trustedPeers) {
		return ip
	}
	if ip := ipFromHeaders(ctx, headers); ip != noIP {
		return ip
	}
	// No ip from the headers, return the peer ip.
	return ip
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}

// UnaryServerInterceptor returns a new unary server interceptor that extracts the real client IP from request headers.
// It checks if the request comes from a trusted peer, and if so, extracts the IP from the configured headers.
// The real IP is added to the request context.
func UnaryServerInterceptor(trustedPeers []netip.Prefix, headers []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ip := getRemoteIP(ctx, trustedPeers, headers)
		if ip != noIP {
			ctx = context.WithValue(ctx, realipKey{}, ip)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that extracts the real client IP from request headers.
// It checks if the request comes from a trusted peer, and if so, extracts the IP from the configured headers.
// The real IP is added to the request context.
func StreamServerInterceptor(trustedPeers []netip.Prefix, headers []string) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ip := getRemoteIP(stream.Context(), trustedPeers, headers)
		if ip != noIP {
			return handler(srv, &serverStream{
				ServerStream: stream,
				ctx:          context.WithValue(stream.Context(), realipKey{}, ip),
			})
		}
		return handler(srv, stream)
	}
}
