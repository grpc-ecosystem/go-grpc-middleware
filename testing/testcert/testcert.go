// Package testcert generates, caches and provides to requestors
// test TLS certificate and key.
package testcert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

var (
	certPEM []byte
	keyPEM  []byte
)

func init() {
	var err error
	certPEM, keyPEM, err = generateCertAndKey([]string{"localhost", "example.com"})
	if err != nil {
		panic("unable to generate test certificate/key: " + err.Error())
	}
}

// CertPEM returns the cached PEM-encoded test TLS certificate.
func CertPEM() []byte {
	return certPEM
}

// KeyPEM returns the cached PEM-encoded key for the TLS certificate.
func KeyPEM() []byte {
	return keyPEM
}

// KeyPairPEM returns the cached PEM-encoded test certificate and key.
// The returned values could be used as input for https://golang.org/pkg/crypto/tls/#X509KeyPair.
func KeyPairPEM() ([]byte, []byte) {
	return CertPEM(), KeyPEM()
}

// generateCertAndKey copied from https://github.com/johanbrandhorst/certify/blob/master/issuers/vault/vault_suite_test.go#L255
// with minor modifications.
func generateCertAndKey(san []string) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "Certify Test Cert",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              san,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return nil, nil, err
	}
	certOut := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	keyOut := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return certOut, keyOut, nil
}
