// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	localnet []net.IPNet = []net.IPNet{
		{
			IP:   net.IPv4(127, 0, 0, 1),
			Mask: net.IPv4Mask(255, 0, 0, 0),
		},
	}

	privatenet []net.IPNet = []net.IPNet{
		{
			IP:   net.IPv4(10, 0, 0, 0),
			Mask: net.IPv4Mask(255, 0, 0, 0),
		},
		{
			IP:   net.IPv4(172, 16, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
		{
			IP:   net.IPv4(192, 168, 0, 0),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		},
	}

	privateIP net.IP = net.IPv4(192, 168, 0, 1)
	publicIP  net.IP = net.IPv4(8, 8, 8, 8)
	localhost net.IP = net.IPv4(127, 0, 0, 1)
)

const (
	xForwardedFor = "x-forwarded-for"
	xRealIP       = "x-real-ip"
	trueClientIP  = "true-client-ip"
)

func localhostPeer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: localhost,
		},
	}
}

func publicPeer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: publicIP,
		},
	}
}

func privatePeer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: privateIP,
		},
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	cases := []struct {
		name         string
		trustedPeers []net.IPNet
		headerKeys   []string
		inputHeaders map[string]string
		peer         *peer.Peer
		expectedIP   net.IP
	}{
		{
			// Test that if there is no peer, we don't get an IP.
			name:         "no peer",
			trustedPeers: localnet,
			headerKeys:   []string{xForwardedFor},
			inputHeaders: map[string]string{
				xForwardedFor: localhost.String(),
			},
			peer:       nil,
			expectedIP: nil,
		},
		{
			// Test that if the remote peer is trusted and the header contains
			// a comma separated list of valid IPs, we get right most one.
			name:         "trusted peer header csv",
			trustedPeers: localnet,
			headerKeys:   []string{xForwardedFor},
			inputHeaders: map[string]string{
				xForwardedFor: fmt.Sprintf("%s,%s", localhost.String(), publicIP.String()),
			},
			peer:       localhostPeer(),
			expectedIP: publicIP,
		},
		{
			// Test that if the remote peer is trusted and the header contains
			// a single valid IP, we get that IP.
			name:         "trusted peer single",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP},
			inputHeaders: map[string]string{
				xRealIP: privateIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: privateIP,
		},
		{
			// Test that if the trusted peers list is larger than 1 network and
			// the remote peer is in the third network, we get the right IP.
			name:         "trusted peer multiple",
			trustedPeers: privatenet,
			headerKeys:   []string{trueClientIP},
			inputHeaders: map[string]string{
				trueClientIP: publicIP.String(),
			},
			peer:       privatePeer(),
			expectedIP: publicIP,
		},
		{
			// Test that if the remote peer is not trusted and the header
			// contains a single valid IP, we get that the peer IP.
			name:         "untrusted peer single",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP},
			inputHeaders: map[string]string{
				xRealIP: privateIP.String(),
			},
			peer:       publicPeer(),
			expectedIP: publicIP,
		},
		{
			// Test that if the peer is trusted and several headers are
			// provided, the interceptor reads the IP from the first header in
			// the list.
			name:         "trusted peer multiple headers",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP, trueClientIP},
			inputHeaders: map[string]string{
				xRealIP:      privateIP.String(),
				trueClientIP: publicIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: privateIP,
		},
		{
			// Test that if the peer is trusted and several headers are
			// configured, but only one is provided, the interceptor reads the
			// IP from the provided header.
			name:         "trusted peer multiple header configured single provided",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP, trueClientIP, xForwardedFor},
			inputHeaders: map[string]string{
				trueClientIP: publicIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: publicIP,
		},
		{
			// Test that if the peer is trusted and several headers are, but no
			// header is provided, the interceptor reads the IP from the peer.
			//
			// This indicates that the proxy is not configured to forward the
			// IP. Using the peer IP is better than nothing.
			name:         "trusted peer multiple header configured none provided",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP, trueClientIP, xForwardedFor},
			peer:         localhostPeer(),
			expectedIP:   localhost,
		},
		{
			// Test that if the peer is not trusted, but several headers are
			// provided, the interceptor reads the IP from peer.
			name:         "untrusted peer multiple headers",
			trustedPeers: nil,
			inputHeaders: map[string]string{
				xRealIP:      privateIP.String(),
				trueClientIP: localhost.String(),
			},
			peer:       publicPeer(),
			expectedIP: publicIP,
		},
		{
			// Test that if the peer is not trusted and several headers are
			// configured, but only one is provided, the interceptor reads the
			// IP from the peer.
			//
			// This is because the peer is untrusted, and as such we cannot
			// trust the headers.
			name:         "untrusted peer multiple header configured single provided",
			trustedPeers: nil,
			headerKeys:   []string{xRealIP, trueClientIP, xForwardedFor},
			inputHeaders: map[string]string{
				trueClientIP: publicIP.String(),
			},
			peer:       publicPeer(),
			expectedIP: publicIP,
		},
		{
			// Test that if the peer is trusted, but the provided headers
			// contain malformed IP addresses, the interceptor reads the IP
			// from the peer.
			name:         "trusted peer malformed header",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP, trueClientIP, xForwardedFor},
			inputHeaders: map[string]string{
				trueClientIP: "malformed",
			},
			peer:       localhostPeer(),
			expectedIP: localhost,
		},
		{
			name:         "",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP},
			peer: &peer.Peer{
				Addr: &net.UnixAddr{Name: "unix", Net: "unix"},
			},
			expectedIP: nil,
		},
		{
			// Test that header casing is ignored.
			name:         "header casing",
			trustedPeers: localnet,
			headerKeys:   []string{xRealIP},
			inputHeaders: map[string]string{
				"X-Real-IP": privateIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: privateIP,
		},
	}

	for _, c := range cases {
		c := c
		t.Run("unary", func(t *testing.T) {
			t.Run(c.name, func(t *testing.T) {
				interceptor := UnaryServerInterceptor(c.trustedPeers, c.headerKeys)
				handler := func(ctx context.Context, req any) (any, error) {
					ip, _ := FromContext(ctx)

					assert.Equal(t, c.expectedIP, ip)
					return nil, nil
				}
				info := &grpc.UnaryServerInfo{
					FullMethod: "FakeMethod",
				}
				ctx := context.Background()
				if c.peer != nil {
					ctx = peer.NewContext(ctx, c.peer)
				}
				if c.inputHeaders != nil {
					md := metadata.New(c.inputHeaders)
					ctx = metadata.NewIncomingContext(ctx, md)
				}

				resp, err := interceptor(ctx, nil, info, handler)
				assert.Nil(t, resp)
				assert.NoError(t, err)
			})
		})
		t.Run("stream", func(t *testing.T) {
			t.Run(c.name, func(t *testing.T) {
				interceptor := StreamServerInterceptor(c.trustedPeers, c.headerKeys)
				handler := func(srv any, stream grpc.ServerStream) error {
					ip, _ := FromContext(stream.Context())

					assert.Equal(t, c.expectedIP, ip)
					return nil
				}
				info := &grpc.StreamServerInfo{
					FullMethod: "FakeMethod",
				}
				ctx := context.Background()
				if c.peer != nil {
					ctx = peer.NewContext(ctx, c.peer)
				}
				if c.inputHeaders != nil {
					md := metadata.New(c.inputHeaders)
					ctx = metadata.NewIncomingContext(ctx, md)
				}

				err := interceptor(nil, &serverStream{ctx: ctx}, info, handler)
				assert.NoError(t, err)
			})
		})
	}
}
