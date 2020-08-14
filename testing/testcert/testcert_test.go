package testcert

import (
	"testing"
)

func TestCertKeyPEM(t *testing.T) {
	cert, key := KeyPairPEM()
	if len(cert) == 0 {
		t.Error("empty cert")
	}
	if len(key) == 0 {
		t.Error("empty key")
	}
}
