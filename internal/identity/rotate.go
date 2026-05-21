// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"context"
	"time"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// RotateOptions describes a Rotate call. The existing keychain content is
// archived to ArchivePath before a fresh key + CSR + .p12 are generated.
// If ArchivePath is empty, no archive is written (use Create directly when
// archival is not wanted).
type RotateOptions struct {
	Subject              CSRSubject
	Keychain             string
	OutPrefix            string // new identity outputs; <prefix>.csr and <prefix>.p12
	P12Password          string // new identity .p12 password (optional)
	ArchivePath          string // existing identities archive; if "" no archive
	ArchivePassword      string // archive .p12 password (optional)
	KeychainUnlockSecret string // env:VAR or keyring: ref for keychain password (issue #11)
}

// RotateResult is the typed outcome.
type RotateResult struct {
	ArchivePath          string       `json:"archive_path,omitempty"`
	ArchivedAt           string       `json:"archived_at,omitempty"`
	ArchiveP12Password   string       `json:"archive_p12_password,omitempty"`
	Created              CreateResult `json:"created"`
}

// Rotate prepares a fresh signing identity while preserving the current
// one as an encrypted PKCS#12 archive.
//
// Sequence:
//   1. If ArchivePath is set, export the current keychain to that path as
//      an encrypted PKCS#12 (security export).
//   2. Generate a new RSA-2048 keypair + CSR + encrypted .p12 backup and
//      import the new key into the same keychain (Create).
//
// Both old and new private keys remain in the keychain afterward — Apple
// allows multiple valid Developer ID certs per team. Use `macforge identity
// list` to confirm which fingerprints are present.
func (s *Service) Rotate(ctx context.Context, opts RotateOptions) (RotateResult, error) {
	if opts.Keychain == "" {
		return RotateResult{}, mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Rotate", "keychain name is required")
	}

	var result RotateResult

	if opts.ArchivePath != "" {
		exp, err := s.Export(ctx, ExportOptions{
			Keychain:    opts.Keychain,
			OutPath:     opts.ArchivePath,
			P12Password: opts.ArchivePassword,
		})
		if err != nil {
			return result, err
		}
		result.ArchivePath = exp.Path
		result.ArchiveP12Password = exp.GeneratedP12Password
		result.ArchivedAt = time.Now().UTC().Format(time.RFC3339)
	}

	created, err := s.Create(ctx, CreateOptions{
		Subject:              opts.Subject,
		Keychain:             opts.Keychain,
		OutPrefix:            opts.OutPrefix,
		P12Password:          opts.P12Password,
		KeychainUnlockSecret: opts.KeychainUnlockSecret,
	})
	if err != nil {
		return result, err
	}
	result.Created = created
	return result, nil
}
