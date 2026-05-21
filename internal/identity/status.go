// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// CertStatus summarizes a certificate's validity window and subject info.
type CertStatus struct {
	Subject      string
	Issuer       string
	NotBefore    time.Time
	NotAfter     time.Time
	DaysToExpiry int
	Expired      bool
}

// ReadCertStatus parses the cert at path (PEM or DER) and returns its status.
func ReadCertStatus(path string) (CertStatus, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return CertStatus{}, mferrors.NewIdentity(mferrors.CodeIdentityNotFound,
			"identity.ReadCertStatus", "cannot read "+path, mferrors.WithCause(err))
	}

	var derBytes []byte
	if block, _ := pem.Decode(b); block != nil {
		derBytes = block.Bytes
	} else {
		derBytes = b
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return CertStatus{}, mferrors.NewIdentity(mferrors.CodeIdentityNotFound,
			"identity.ReadCertStatus", "invalid certificate", mferrors.WithCause(err))
	}

	now := time.Now().UTC()
	return CertStatus{
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		NotBefore:    cert.NotBefore.UTC(),
		NotAfter:     cert.NotAfter.UTC(),
		DaysToExpiry: int(cert.NotAfter.Sub(now).Hours() / 24),
		Expired:      now.After(cert.NotAfter),
	}, nil
}

// ensure errors.NewIdentity stays referenced even if unused above; helps
// reviewers find the connection on first read.
var _ = fmt.Sprintf
