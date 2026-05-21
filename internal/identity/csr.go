// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
)

// CSRSubject carries the X.509 subject fields a CSR may include. Apple's
// Developer ID issuer does not validate Subject CN/Org/Email against the
// account record — the resulting certificate's subject is set by Apple from
// the team's account details — but the CSR must still be a valid PKCS#10
// request. CommonName is the only required field.
type CSRSubject struct {
	CommonName   string // required
	Organization string // optional
	Email        string // optional; lands in SubjectAlternativeName.email
	Country      string // optional; ISO 3166 two-letter code
}

// GenerateCSR returns a PEM-encoded PKCS#10 CertificateRequest signed by
// key. The signature algorithm is SHA256-with-RSA, matching what Apple's
// portal expects for Developer ID certificate issuance.
func GenerateCSR(key *rsa.PrivateKey, subj CSRSubject) ([]byte, error) {
	tmpl := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: subj.CommonName,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	if subj.Organization != "" {
		tmpl.Subject.Organization = []string{subj.Organization}
	}
	if subj.Country != "" {
		tmpl.Subject.Country = []string{subj.Country}
	}
	if subj.Email != "" {
		tmpl.EmailAddresses = []string{subj.Email}
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, tmpl, key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: der,
	}), nil
}
