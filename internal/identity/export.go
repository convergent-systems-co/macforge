// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// ExportOptions describes an Export call. Empty P12Password means a fresh
// random passphrase is generated and surfaced in the result.
type ExportOptions struct {
	Keychain    string // source keychain (must exist)
	OutPath     string // .p12 destination
	P12Password string // optional; generated if empty
}

// ExportResult is the typed outcome of Export.
type ExportResult struct {
	Path                 string `json:"path"`
	Keychain             string `json:"keychain"`
	GeneratedP12Password string `json:"generated_p12_password,omitempty"`
}

// Export writes all identities (cert + private key pairs) from the named
// keychain to an AES-encrypted PKCS#12 file at OutPath, sealed with
// P12Password. Wraps `security export`. Common.md §4: the password is
// passed only to `security` and (when generated) returned ONCE in the
// result — never persisted by MacForge.
func (s *Service) Export(ctx context.Context, opts ExportOptions) (ExportResult, error) {
	if opts.Keychain == "" {
		return ExportResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Export", "keychain name is required")
	}
	if opts.OutPath == "" {
		return ExportResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Export", "OutPath is required")
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutPath), 0o755); err != nil {
		return ExportResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Export", "mkdir for export path failed", mferrors.WithCause(err))
	}

	pw := opts.P12Password
	generated := ""
	if pw == "" {
		buf := make([]byte, 24)
		if _, err := rand.Read(buf); err != nil {
			return ExportResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
				"identity.Export", "random password generation failed", mferrors.WithCause(err))
		}
		pw = base64.RawURLEncoding.EncodeToString(buf)
		generated = pw
	}

	if err := s.sec.Export(ctx, opts.Keychain, opts.OutPath, pw); err != nil {
		return ExportResult{}, err
	}

	// security export writes 0644 by default. The file contains an encrypted
	// private key — tighten the on-disk perms regardless.
	_ = os.Chmod(opts.OutPath, 0o600)

	return ExportResult{
		Path:                 opts.OutPath,
		Keychain:             opts.Keychain,
		GeneratedP12Password: generated,
	}, nil
}
