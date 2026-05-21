// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"

	"software.sslmate.com/src/go-pkcs12"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// CreateOptions describes a Create call. OutPrefix is a base path:
// the CSR lands at "<OutPrefix>.csr", the encrypted PKCS#12 backup at
// "<OutPrefix>.p12". P12Password may be empty — in that case a random
// passphrase is generated and returned in the result for the caller to
// surface to the user once.
type CreateOptions struct {
	Subject     CSRSubject
	Keychain    string // target macforge keychain (must already exist)
	OutPrefix   string // file path prefix; .csr and .p12 are appended
	P12Password string // optional; if empty, a random passphrase is generated
}

// CreateResult is the typed outcome of Create.
type CreateResult struct {
	CSRPath               string `json:"csr_path"`
	P12Path               string `json:"p12_path"`
	GeneratedP12Password  string `json:"generated_p12_password,omitempty"`
	Keychain              string `json:"keychain"`
	PublicKeyFingerprint  string `json:"public_key_fingerprint"`
	CommonName            string `json:"common_name"`
}

// Create generates a new RSA-2048 keypair, writes the PEM-encoded PKCS#10
// CSR to "<OutPrefix>.csr", writes an encrypted PKCS#12 (key + self-signed
// filler cert) to "<OutPrefix>.p12", and imports the same .p12 into the
// macforge keychain so signing works the moment Apple issues the real cert.
//
// Per Common.md §4 the private key never lives on disk in PEM form: the
// PKCS#12 is encrypted, and the keychain import is via the same encrypted
// bundle. If P12Password is empty, a fresh URL-safe-base64 passphrase is
// generated and returned in CreateResult.GeneratedP12Password.
func (s *Service) Create(ctx context.Context, opts CreateOptions) (CreateResult, error) {
	if opts.Subject.CommonName == "" {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "Subject.CommonName is required")
	}
	if opts.Keychain == "" {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "target keychain name is required")
	}
	if opts.OutPrefix == "" {
		opts.OutPrefix = "./identity"
	}

	key, err := GenerateRSAKey()
	if err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "rsa keygen failed", mferrors.WithCause(err))
	}

	csr, err := GenerateCSR(key, opts.Subject)
	if err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "csr generation failed", mferrors.WithCause(err))
	}

	csrPath := opts.OutPrefix + ".csr"
	if err := os.MkdirAll(filepath.Dir(csrPath), 0o755); err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "mkdir for csr failed", mferrors.WithCause(err))
	}
	if err := os.WriteFile(csrPath, csr, 0o644); err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "writing csr failed", mferrors.WithCause(err))
	}

	filler, err := newFillerCert(key, opts.Subject.CommonName)
	if err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "filler cert generation failed", mferrors.WithCause(err))
	}

	pw := opts.P12Password
	generated := ""
	if pw == "" {
		// 24 bytes → 32-char URL-safe base64. Plenty of entropy.
		buf := make([]byte, 24)
		if _, err := rand.Read(buf); err != nil {
			return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
				"identity.Create", "random password generation failed", mferrors.WithCause(err))
		}
		pw = base64.RawURLEncoding.EncodeToString(buf)
		generated = pw
	}

	p12Bytes, err := pkcs12.Modern.Encode(key, filler, nil, pw)
	if err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "pkcs12 encode failed", mferrors.WithCause(err))
	}

	p12Path := opts.OutPrefix + ".p12"
	if err := os.WriteFile(p12Path, p12Bytes, 0o600); err != nil {
		return CreateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Create", "writing pkcs12 failed", mferrors.WithCause(err))
	}

	if err := s.sec.Import(ctx, p12Path, opts.Keychain, pw); err != nil {
		return CreateResult{}, err
	}

	return CreateResult{
		CSRPath:              csrPath,
		P12Path:              p12Path,
		GeneratedP12Password: generated,
		Keychain:             opts.Keychain,
		PublicKeyFingerprint: publicKeyFingerprint(filler),
		CommonName:           opts.Subject.CommonName,
	}, nil
}

// publicKeyFingerprint returns the SHA-256 of the cert's
// SubjectPublicKeyInfo as lowercase hex (full 64 chars). The same public
// key is bound to the filler cert and the eventually-issued Apple cert,
// so this fingerprint is stable across the cert-import step.
func publicKeyFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	sum := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(sum[:])
}
