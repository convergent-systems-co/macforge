// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/convergent-systems-co/macforge/internal/identity"
)

func writeTestCert(t *testing.T, dir string, notBefore, notAfter time.Time) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Developer ID Application: ACME (XYZ1234567)"},
		Issuer:       pkix.Name{CommonName: "Apple Worldwide Developer Relations CA"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	path := filepath.Join(dir, "cert.pem")
	f, _ := os.Create(path)
	defer f.Close()
	_ = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	return path
}

func TestReadCertStatus_NotExpired(t *testing.T) {
	dir := t.TempDir()
	path := writeTestCert(t,
		dir,
		time.Now().Add(-24*time.Hour),
		time.Now().Add(365*24*time.Hour),
	)
	st, err := identity.ReadCertStatus(path)
	if err != nil {
		t.Fatalf("ReadCertStatus: %v", err)
	}
	if st.Expired {
		t.Fatal("Expired = true, want false")
	}
	if st.DaysToExpiry < 360 || st.DaysToExpiry > 366 {
		t.Fatalf("DaysToExpiry = %d, want ~365", st.DaysToExpiry)
	}
}

func TestReadCertStatus_Expired(t *testing.T) {
	dir := t.TempDir()
	path := writeTestCert(t,
		dir,
		time.Now().Add(-2*365*24*time.Hour),
		time.Now().Add(-24*time.Hour),
	)
	st, err := identity.ReadCertStatus(path)
	if err != nil {
		t.Fatalf("ReadCertStatus: %v", err)
	}
	if !st.Expired {
		t.Fatal("Expired = false, want true")
	}
}
