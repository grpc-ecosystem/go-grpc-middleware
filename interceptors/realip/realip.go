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
	vals := metadata.ValueFromIncomingContext(ctx, key)

	if len(vals) == 0 {
		return ""
	}

	return vals[0]
}

func ipFromXForwardedFoR(trustedProxies []netip.Prefix, ips []string, idx int) netip.Addr {
	for i := idx; i >= 0; i-- {
		h := strings.TrimSpace(ips[i])
		ip, err := netip.ParseAddr(h)
		if err != nil {
			return noIP
		}
		if !ipInNets(ip, trustedProxies) {
			return ip
		}
	}
	return noIP
}

func ipFromHeaders(ctx context.Context, headers []string, trustedProxies []netip.Prefix, trustedProxyCnt uint) netip.Addr {
	for _, header := range headers {
		a := strings.Split(getHeader(ctx, header), ",")
		idx := len(a) - 1
		if header == XForwardedFor {
			idx -= int(trustedProxyCnt)
			if idx < 0 {
				continue
			}
			return ipFromXForwardedFoR(trustedProxies, a, idx)
		}
		h := strings.TrimSpace(a[idx])
		ip, err := netip.ParseAddr(h)
		if err == nil {
			return ip
		}
	}
	return noIP
}

func getRemoteIP(ctx context.Context, trustedPeers, trustedProxies []netip.Prefix, headers []string, proxyCnt uint) netip.Addr {
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
	if resolvedIP := ipFromHeaders(ctx, headers, trustedProxies, proxyCnt); resolvedIP != noIP {
		return resolvedIP
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
// See UnaryServerInterceptorOpts as it allows to configure trusted proxy ips list and count that should work better with Google LB
func UnaryServerInterceptor(trustedPeers []netip.Prefix, headers []string) grpc.UnaryServerInterceptor {
	return UnaryServerInterceptorOpts(WithTrustedPeers(trustedPeers), WithHeaders(headers))
}

// StreamServerInterceptor returns a new stream server interceptor that extracts the real client IP from request headers.
// It checks if the request comes from a trusted peer, and if so, extracts the IP from the configured headers.
// The real IP is added to the request context.
// See UnaryServerInterceptorOpts as it allows to configure trusted proxy ips list and count that should work better with Google LB
func StreamServerInterceptor(trustedPeers []netip.Prefix, headers []string) grpc.StreamServerInterceptor {
	return StreamServerInterceptorOpts(WithTrustedPeers(trustedPeers), WithHeaders(headers))
}

// UnaryServerInterceptorOpts returns a new unary server interceptor that extracts the real client IP from request headers.
// It checks if the request comes from a trusted peer, validates headers against trusted proxies list and trusted proxies count
// then it extracts the IP from the configured headers.
// The real IP is added to the request context.
func UnaryServerInterceptorOpts(opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOpts(opts)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ip := getRemoteIP(ctx, o.trustedPeers, o.trustedProxies, o.headers, o.trustedProxiesCount)
		if ip != noIP {
			ctx = context.WithValue(ctx, realipKey{}, ip)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptorOpts returns a new stream server interceptor that extracts the real client IP from request headers.
// It checks if the request comes from a trusted peer, validates headers against trusted proxies list and trusted proxies count
// then it extracts the IP from the configured headers.
// The real IP is added to the request context.
func StreamServerInterceptorOpts(opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOpts(opts)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ip := getRemoteIP(stream.Context(), o.trustedPeers, o.trustedProxies, o.headers, o.trustedProxiesCount)
		if ip != noIP {
			return handler(srv, &serverStream{
				ServerStream: stream,
				ctx:          context.WithValue(stream.Context(), realipKey{}, ip),
			})
		}
		return handler(srv, stream)
	}
}
