// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity_test

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/identity"
)

func TestGenerateCSR_RoundTrip(t *testing.T) {
	key, err := identity.GenerateRSAKey()
	if err != nil {
		t.Fatalf("GenerateRSAKey: %v", err)
	}

	subj := identity.CSRSubject{
		CommonName:   "Convergent Systems Co.",
		Organization: "Convergent Systems Co.",
		Email:        "you@example.com",
		Country:      "US",
	}
	pemBytes, err := identity.GenerateCSR(key, subj)
	if err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Fatalf("PEM decode failed; output:\n%s", string(pemBytes))
	}
	if block.Type != "CERTIFICATE REQUEST" {
		t.Fatalf("PEM type = %q, want CERTIFICATE REQUEST", block.Type)
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		t.Fatalf("ParseCertificateRequest: %v", err)
	}
	if csr.Subject.CommonName != "Convergent Systems Co." {
		t.Fatalf("CN = %q, want Convergent Systems Co.", csr.Subject.CommonName)
	}
	if len(csr.Subject.Organization) == 0 || csr.Subject.Organization[0] != "Convergent Systems Co." {
		t.Fatalf("Org = %v, want [Convergent Systems Co.]", csr.Subject.Organization)
	}
	if len(csr.Subject.Country) == 0 || csr.Subject.Country[0] != "US" {
		t.Fatalf("Country = %v, want [US]", csr.Subject.Country)
	}
	if len(csr.EmailAddresses) == 0 || csr.EmailAddresses[0] != "you@example.com" {
		t.Fatalf("Email = %v, want [you@example.com]", csr.EmailAddresses)
	}
	if csr.SignatureAlgorithm != x509.SHA256WithRSA {
		t.Fatalf("SignatureAlgorithm = %v, want SHA256-RSA", csr.SignatureAlgorithm)
	}
	if err := csr.CheckSignature(); err != nil {
		t.Fatalf("CSR signature check failed: %v", err)
	}
}

func TestGenerateCSR_MinimalSubject(t *testing.T) {
	key, _ := identity.GenerateRSAKey()
	pemBytes, err := identity.GenerateCSR(key, identity.CSRSubject{CommonName: "Only CN"})
	if err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}
	block, _ := pem.Decode(pemBytes)
	csr, _ := x509.ParseCertificateRequest(block.Bytes)
	if csr.Subject.CommonName != "Only CN" {
		t.Fatalf("CN = %q", csr.Subject.CommonName)
	}
}
