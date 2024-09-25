// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

/*
Package realip is a middleware that extracts the real IP of requests based on
header values.

The real IP is subsequently placed inside the context of each request and can
be retrieved using the [FromContext] function.

The middleware is designed to work with gRPC servers serving clients over
TCP/IP connections. If no headers are found, the middleware will return the
remote peer address as the real IP. If remote peer address is not a TCP/IP
address, the middleware will return nil as the real IP.

Headers provided by clients in the request will be searched for in the order
of the list provided to the middleware. The first header that contains a valid
IP address will be used as the real IP.

Comma separated headers are supported. The last, rightmost, IP address in the
header will be used as the real IP.

# Security

There are 2 main security concerns when deriving the real IP from request
headers:

 1. Risk of spoofing the real IP by setting a header value.
 2. Risk of injecting a header value that causes a denial of service.

To mitigate the risk of spoofing, the middleware introduces the concept of
"trusted peers". Trusted peers are defined as a list of IP networks that are
verified by the gRPC server operator to be trusted. If the peer address is found
to be within one of the trusted networks, the middleware will attempt to extract
the real IP from the request headers. If the peer address is not found to be
within one of the trusted networks, the peer address will be returned as the
real IP.

"trusted peer" in this context means that the peer is configured to overwrite the
header value with the real IP. This is typically done by a proxy or load
balancer that is configured to forward the real IP of the client in a header
value. Alternatively, the peer may be configured to append the real IP to the
header value. In this case, the middleware will use the last, rightmost, IP
address in the header as the real IP. Most load balancers, such as NGINX, AWS
ELB, are configured to append the real IP to the header value as their default action.
However, Google Cloud Load Balancer for `X-Forwarded-For` follows the pattern:
`<client-ip>,<load-balancer-ip>`. Hence we need to have an ability to exact the
real ip from the header ignoring the LB/proxy IPs.

### Supported Methods for Extracting Real IP:

This is based on
[Selecting an IP address](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For#selecting_an_ip_address).

 1. Trusted Proxy Count

With this method, the count of reverse proxies between the internet and the server is configured.
The middleware searches the `X-Forwarded-For` IP list from the rightmost by that count.

 2. Trusted Proxy List

Alternatively, you can configure a list of trusted reverse proxies by specifying their
IPs or IP ranges. The middleware will then search the `X-Forwarded-For` IP list from
the rightmost, skipping all addresses that are on the trusted proxy list.
The first non-matching address is considered the target address.

# Individual IP addresses as trusted peers

When creating the list of trusted peers, it is possible to specify individual IP
addresses. This is useful when your proxy or load balancer has a set of
well-known addresses.

The following example shows how to specify individual IP addresses as trusted
peers:

	trusted := []net.IPNet{
	    {IP: net.IPv4(192, 168, 0, 1), Mask: net.IPv4Mask(255, 255, 255, 255)},
	    {IP: net.IPv4(192, 168, 0, 2), Mask: net.IPv4Mask(255, 255, 255, 255)},
	}

In the above example, the middleware will only attempt to extract the real IP
from the request headers if the peer address is either 192.168.0.1 or
192.168.0.2.

# Headers

Headers to search for are specified as a list of strings when creating the
middleware. The middleware will search for the headers in the order specified
and use the first header that contains a valid IP address as the real IP.

The following are examples of headers that may contain the real IP:

  - X-Forwarded-For: This header is set by proxies and contains a comma
    separated list of IP addresses. Each proxy that forwards the request will
    append the real IP to the header value.
  - X-Real-IP: This header is set by NGINX and contains the real IP as a string
    containing a single IP address.
  - Forwarded-For: Header defined by RFC7239. This header is set by proxies and
    contains the real IP as a string containing a single IP address. Please note
    that the obfuscated identifier from section 6.3 of RFC7239, and that the
    unknown identifier from section 6.2 of RFC7239 are not supported.
  - True-Client-IP: This header is set by Cloudflare and contains the real IP
    as a string containing a single IP address.

# Usage

Please see examples for simple examples of use.
*/
package realip
