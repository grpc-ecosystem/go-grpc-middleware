// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package realip

import "net/netip"

// options represents the configuration options for the realip middleware.
type options struct {
	// trustedPeers is a list of trusted peers network prefixes.
	trustedPeers []netip.Prefix

	// trustedProxies is a list of trusted proxies network prefixes.
	// The first rightmost non-matching IP when going through X-Forwarded-For is considered the client IP.
	trustedProxies []netip.Prefix

	// trustedProxiesCount specifies the number of proxies in front that may append X-Forwarded-For.
	// It defaults to 0.
	trustedProxiesCount uint

	// headers specifies the headers to use in real IP extraction when the request is from a trusted peer.
	headers []string
}

// An Option lets you add options to realip interceptors using With* functions.
type Option func(*options)

func evaluateOpts(opts []Option) *options {
	optCopy := &options{}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// WithTrustedPeers sets the trusted peers network prefixes.
func WithTrustedPeers(peers []netip.Prefix) Option {
	return func(o *options) {
		o.trustedPeers = peers
	}
}

// WithTrustedProxies sets the trusted proxies network prefixes.
func WithTrustedProxies(proxies []netip.Prefix) Option {
	return func(o *options) {
		o.trustedProxies = proxies
	}
}

// WithTrustedProxiesCount sets the number of trusted proxies that may append X-Forwarded-For.
func WithTrustedProxiesCount(count uint) Option {
	return func(o *options) {
		o.trustedProxiesCount = count
	}
}

// WithHeaders sets the headers to use in real IP extraction for requests from trusted peers.
func WithHeaders(headers []string) Option {
	return func(o *options) {
		o.headers = headers
	}
}
