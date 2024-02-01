// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	localnet []netip.Prefix = []netip.Prefix{
		netip.MustParsePrefix("127.0.0.1/8"),
		netip.MustParsePrefix("::1/128"),
	}

	privatenet []netip.Prefix = []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("2002:c0a8::/32"),
	}

	privateIP  netip.Addr = netip.MustParseAddr("192.168.0.1")
	privateIP6 netip.Addr = netip.MustParseAddr("::ffff:c0a8:1")
	publicIP   netip.Addr = netip.MustParseAddr("8.8.8.8")
	publicIP6  netip.Addr = netip.MustParseAddr("::ffff:808:808")
	localhost  netip.Addr = netip.MustParseAddr("127.0.0.1")
	localhost6 netip.Addr = netip.MustParseAddr("::1")
)

func localhostPeer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: net.ParseIP(localhost.String()),
		},
	}
}

func localhost6Peer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: net.ParseIP(localhost6.String()),
		},
	}
}

func publicPeer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: net.ParseIP(publicIP.String()),
		},
	}
}

func privatePeer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: net.ParseIP(privateIP.String()),
		},
	}
}

func private6Peer() *peer.Peer {
	return &peer.Peer{
		Addr: &net.TCPAddr{
			IP: net.ParseIP(privateIP6.String()),
		},
	}
}

type testCase struct {
	trustedPeers []netip.Prefix
	headerKeys   []string
	inputHeaders map[string]string
	peer         *peer.Peer
	expectedIP   netip.Addr
}

func testUnaryServerInterceptor(t *testing.T, c testCase) {
	interceptor := UnaryServerInterceptor(c.trustedPeers, c.headerKeys)
	handler := func(ctx context.Context, req any) (any, error) {
		ip, _ := FromContext(ctx)
		fmt.Println(ip)
		fmt.Println(c.expectedIP)

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
}

func testStreamServerInterceptor(t *testing.T, c testCase) {
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
}

func TestInterceptor(t *testing.T) {
	t.Run("no peer", func(t *testing.T) {
		tc := testCase{
			// Test that if there is no peer, we don't get an IP.
			trustedPeers: localnet,
			headerKeys:   []string{XForwardedFor},
			inputHeaders: map[string]string{
				XForwardedFor: localhost.String(),
			},
			peer:       nil,
			expectedIP: netip.Addr{},
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})

	t.Run("trusted peer header csv", func(t *testing.T) {
		tc := testCase{
			// Test that if the remote peer is trusted and the header contains
			// a comma separated list of valid IPs, we get right most one.
			trustedPeers: localnet,
			headerKeys:   []string{XForwardedFor},
			inputHeaders: map[string]string{
				XForwardedFor: fmt.Sprintf("%s,%s", localhost.String(), publicIP.String()),
			},
			peer:       localhostPeer(),
			expectedIP: publicIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("trusted peer single", func(t *testing.T) {
		tc := testCase{
			// Test that if the remote peer is trusted and the header contains
			// a single valid IP, we get that IP.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp},
			inputHeaders: map[string]string{
				XRealIp: privateIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: privateIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("trusted peer multiple", func(t *testing.T) {
		tc := testCase{
			// Test that if the trusted peers list is larger than 1 network and
			// the remote peer is in the third network, we get the right IP.
			trustedPeers: privatenet,
			headerKeys:   []string{TrueClientIp},
			inputHeaders: map[string]string{
				TrueClientIp: publicIP.String(),
			},
			peer:       privatePeer(),
			expectedIP: publicIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("untrusted peer single", func(t *testing.T) {
		tc := testCase{
			// Test that if the remote peer is not trusted and the header
			// contains a single valid IP, we get that the peer IP.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp},
			inputHeaders: map[string]string{
				XRealIp: privateIP.String(),
			},
			peer:       publicPeer(),
			expectedIP: publicIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("trusted peer multiple headers", func(t *testing.T) {
		tc := testCase{
			// Test that if the peer is trusted and several headers are
			// provided, the interceptor reads the IP from the first header in
			// the list.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp, TrueClientIp},
			inputHeaders: map[string]string{
				XRealIp:      privateIP.String(),
				TrueClientIp: publicIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: privateIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("trusted peer multiple header configured single provided", func(t *testing.T) {
		tc := testCase{
			// Test that if the peer is trusted and several headers are
			// configured, but only one is provided, the interceptor reads the
			// IP from the provided header.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp, TrueClientIp, XForwardedFor},
			inputHeaders: map[string]string{
				TrueClientIp: publicIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: publicIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("trusted peer multiple header configured none provided", func(t *testing.T) {
		tc := testCase{
			// Test that if the peer is trusted and several headers are, but no
			// header is provided, the interceptor reads the IP from the peer.
			//
			// This indicates that the proxy is not configured to forward the
			// IP. Using the peer IP is better than nothing.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp, TrueClientIp, XForwardedFor},
			peer:         localhostPeer(),
			expectedIP:   localhost,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("untrusted peer multiple headers", func(t *testing.T) {
		tc := testCase{
			// Test that if the peer is not trusted, but several headers are
			// provided, the interceptor reads the IP from peer.
			trustedPeers: nil,
			inputHeaders: map[string]string{
				XRealIp:      privateIP.String(),
				TrueClientIp: localhost.String(),
			},
			peer:       publicPeer(),
			expectedIP: publicIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("untrusted peer multiple header configured single provided", func(t *testing.T) {
		tc := testCase{
			// Test that if the peer is not trusted and several headers are
			// configured, but only one is provided, the interceptor reads the
			// IP from the peer.
			//
			// This is because the peer is untrusted, and as such we cannot
			// trust the headers.
			trustedPeers: nil,
			headerKeys:   []string{XRealIp, TrueClientIp, XForwardedFor},
			inputHeaders: map[string]string{
				TrueClientIp: publicIP.String(),
			},
			peer:       publicPeer(),
			expectedIP: publicIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("trusted peer malformed header", func(t *testing.T) {
		tc := testCase{
			// Test that if the peer is trusted, but the provided headers
			// contain malformed IP addresses, the interceptor reads the IP
			// from the peer.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp, TrueClientIp, XForwardedFor},
			inputHeaders: map[string]string{
				TrueClientIp: "malformed",
			},
			peer:       localhostPeer(),
			expectedIP: localhost,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("ipv6 from grpc peer", func(t *testing.T) {
		tc := testCase{
			trustedPeers: localnet,
			headerKeys:   []string{},
			peer:         localhost6Peer(),
			expectedIP:   localhost6,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("ipv6 from header", func(t *testing.T) {
		tc := testCase{
			trustedPeers: privatenet,
			headerKeys:   []string{XForwardedFor},
			inputHeaders: map[string]string{
				XForwardedFor: publicIP6.String(),
			},
			peer:       private6Peer(),
			expectedIP: publicIP6,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("unix", func(t *testing.T) {
		tc := testCase{
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp},
			peer: &peer.Peer{
				Addr: &net.UnixAddr{Name: "unix", Net: "unix"},
			},
			expectedIP: netip.Addr{},
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
	t.Run("header casing", func(t *testing.T) {
		tc := testCase{
			// Test that header casing is ignored.
			trustedPeers: localnet,
			headerKeys:   []string{XRealIp},
			inputHeaders: map[string]string{
				"X-Real-IP": privateIP.String(),
			},
			peer:       localhostPeer(),
			expectedIP: privateIP,
		}
		t.Run("unary", func(t *testing.T) {
			testUnaryServerInterceptor(t, tc)
		})
		t.Run("stream", func(t *testing.T) {
			testStreamServerInterceptor(t, tc)
		})
	})
}
