#!/bin/bash
# Regenerate the self-signed certificate for local host.

openssl req -x509 -sha256 -nodes -newkey rsa:2048 -days 3650 -keyout localhost.key -out localhost.crt

cat <<EOF > localhost.go
package certs

var LocalhostKey []byte = []byte(\`$(cat localhost.key)\`)

var LocalhostCert []byte = []byte(\`$(cat localhost.crt)\`)
EOF
