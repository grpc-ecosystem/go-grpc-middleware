// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type realipKey struct{}

func FromContext(ctx context.Context) (net.IP, bool) {
	ip, ok := ctx.Value(realipKey{}).(net.IP)
	return ip, ok
}

func remotePeer(ctx context.Context) net.Addr {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	return pr.Addr
}

func ipInNets(ip net.IP, nets []net.IPNet) bool {
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

	if md[key] == nil {
		return ""
	}

	return md[key][0]
}

func ipFromHeaders(ctx context.Context, headers []string) net.IP {
	for _, header := range headers {
		a := strings.Split(getHeader(ctx, header), ",")
		h := strings.TrimSpace(a[len(a)-1])
		if ip := net.ParseIP(h); ip != nil {
			return ip
		}
	}
	return nil
}

func getRemoteIP(ctx context.Context, trustedPeers []net.IPNet, headers []string) net.IP {
	pr := remotePeer(ctx)
	if pr == nil {
		return nil
	}
	strIp, _, err := net.SplitHostPort(pr.String())
	if err != nil {
		return nil
	}
	ip := net.ParseIP(strIp)
	if len(trustedPeers) == 0 || !ipInNets(ip, trustedPeers) {
		return ip
	}
	if ip := ipFromHeaders(ctx, headers); ip != nil {
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

// UnaryServerInterceptor returns a new unary server interceptors that performs request rate limiting.
func UnaryServerInterceptor(trustedPeers []net.IPNet, headers []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ip := getRemoteIP(ctx, trustedPeers, headers)
		if ip != nil {
			ctx = context.WithValue(ctx, realipKey{}, ip)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that performs rate limiting on the request.
func StreamServerInterceptor(trustedPeers []net.IPNet, headers []string) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ip := getRemoteIP(stream.Context(), trustedPeers, headers)
		if ip != nil {
			return handler(srv, &serverStream{
				ServerStream: stream,
				ctx:          context.WithValue(stream.Context(), realipKey{}, ip),
			})
		}
		return handler(srv, stream)
	}
}
