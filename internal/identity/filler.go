// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// newFillerCert returns a self-signed X.509 certificate bound to key. It
// exists only to satisfy PKCS#12 encoders that require a non-nil cert and
// `security import` which expects a cert+key pair. It is replaced (or
// ignored) once Apple issues the real Developer ID certificate, which
// binds to the same private key by matching public key.
//
// The filler cert has NO ExtKeyUsage entries so it will NOT appear in
// `security find-identity -p codesigning` output, keeping it out of
// MacForge's identity-discovery flow.
func newFillerCert(key *rsa.PrivateKey, cn string) (*x509.Certificate, error) {
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: cn + " (macforge bootstrap)"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		// No ExtKeyUsage — keeps it out of codesigning lookups.
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(der)
}
