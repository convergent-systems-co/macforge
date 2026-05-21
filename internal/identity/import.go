// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"context"
	"os"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// ImportOptions describes one cert import.
type ImportOptions struct {
	File        string // path to .cer, .pem, or .p12
	Keychain    string // target keychain name
	P12Password string // empty for .cer/.pem
}

// Import installs a cert or PKCS#12 bundle into the keychain.
func (s *Service) Import(ctx context.Context, opts ImportOptions) error {
	if _, err := os.Stat(opts.File); err != nil {
		return mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"identity.Import",
			"file not found: "+opts.File,
			mferrors.WithCause(err))
	}
	return s.sec.Import(ctx, opts.File, opts.Keychain, opts.P12Password)
}
