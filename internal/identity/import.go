// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"context"
	"os"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
	"github.com/convergent-systems-co/macforge/internal/keychain"
)

// ImportOptions describes one cert import.
type ImportOptions struct {
	File                 string // path to .cer, .pem, or .p12
	Keychain             string // target keychain name
	P12Password          string // empty for .cer/.pem
	KeychainUnlockSecret string // env:VAR or keyring: ref for the KEYCHAIN master password;
	                            // required to grant codesign access to any imported keys
}

// Import installs a cert or PKCS#12 bundle into the keychain. If the import
// is a .p12 (i.e. carries a private key), the partition-list ACL is set so
// the key is usable for non-interactive codesigning. Issue #11.
func (s *Service) Import(ctx context.Context, opts ImportOptions) error {
	if _, err := os.Stat(opts.File); err != nil {
		return mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Import",
			"file not found: "+opts.File,
			mferrors.WithCause(err))
	}
	if err := s.sec.Import(ctx, opts.File, opts.Keychain, opts.P12Password); err != nil {
		return err
	}
	// .cer imports don't add keys; .p12 imports do. Setting the partition
	// list is idempotent and cheap — apply it unconditionally so newly
	// added keys are immediately usable for codesign.
	if opts.KeychainUnlockSecret != "" {
		kcPW, err := keychain.ResolveSecret(opts.KeychainUnlockSecret)
		if err != nil {
			return err
		}
		if err := s.sec.SetKeyPartitionList(ctx, opts.Keychain, kcPW, defaultPartitionList); err != nil {
			return err
		}
	}
	return nil
}

// defaultPartitionList is the partition-list ACL value that grants
// `security` and `codesign` (both Apple-signed tools) non-interactive
// access to private keys in the keychain. Matches the convention used by
// fastlane, xcrun match, and github-actions/import-codesigning-certs.
const defaultPartitionList = "apple-tool:,apple:"
