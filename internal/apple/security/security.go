// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package security wraps the macOS `security` CLI for the operations
// MacForge needs: keychain lifecycle, identity discovery, certificate
// import. Every call flows through the parent apple.Runner.
package security

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/apple"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// Client wraps a Runner for the `security` binary.
type Client struct {
	r apple.Runner
}

// New returns a Client that drives r.
func New(r apple.Runner) *Client { return &Client{r: r} }

// CreateKeychain creates a new keychain file at the default location.
// password is masked from the audit log via the Invocation's Redact list.
func (c *Client) CreateKeychain(ctx context.Context, name, password string) error {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool:   "security",
		Args:   argsCreateKeychain(name, password),
		Redact: []string{password},
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewKeychain(mferrors.CodeKeychainExists,
			"security.CreateKeychain",
			fmt.Sprintf("security create-keychain failed: %s", strings.TrimSpace(string(res.Stderr))),
			mferrors.WithDetails(map[string]any{"keychain": name, "exit": res.ExitCode}))
	}
	return nil
}

// UnlockKeychain unlocks a keychain for the current login session.
func (c *Client) UnlockKeychain(ctx context.Context, name, password string) error {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool:   "security",
		Args:   argsUnlockKeychain(name, password),
		Redact: []string{password},
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewKeychain(mferrors.CodeKeychainLocked,
			"security.UnlockKeychain",
			fmt.Sprintf("security unlock-keychain failed: %s", strings.TrimSpace(string(res.Stderr))),
			mferrors.WithDetails(map[string]any{"keychain": name}))
	}
	return nil
}

// DeleteKeychain removes a keychain file. Refuses to touch login.keychain.
func (c *Client) DeleteKeychain(ctx context.Context, name string) error {
	if strings.HasPrefix(name, "login.") || name == "login.keychain" || name == "login.keychain-db" {
		return mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.DeleteKeychain",
			"refusing to delete login.keychain (security policy)",
			mferrors.WithDetails(map[string]any{"keychain": name}))
	}
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "security",
		Args: []string{"delete-keychain", name},
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.DeleteKeychain",
			strings.TrimSpace(string(res.Stderr)),
			mferrors.WithDetails(map[string]any{"keychain": name}))
	}
	return nil
}

// SetSettings configures auto-lock behavior. lockOnSleep enables -l.
// timeoutSecs sets -t; 0 disables the timeout.
func (c *Client) SetSettings(ctx context.Context, name string, lockOnSleep bool, timeoutSecs int) error {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "security",
		Args: argsSetKeychainSettings(name, lockOnSleep, timeoutSecs),
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.SetSettings",
			strings.TrimSpace(string(res.Stderr)),
			mferrors.WithDetails(map[string]any{"keychain": name}))
	}
	return nil
}

// argv builders — exported only for tests in this package.

func argsCreateKeychain(name, password string) []string {
	return []string{"create-keychain", "-p", password, name}
}

func argsUnlockKeychain(name, password string) []string {
	return []string{"unlock-keychain", "-p", password, name}
}

func argsSetKeychainSettings(name string, lockOnSleep bool, timeoutSecs int) []string {
	out := []string{"set-keychain-settings"}
	if lockOnSleep {
		out = append(out, "-l")
	}
	if timeoutSecs > 0 {
		out = append(out, "-t", strconv.Itoa(timeoutSecs))
	}
	return append(out, name)
}

func argsFindIdentity(keychain, policy string) []string {
	return []string{"find-identity", "-p", policy, "-v", keychain}
}

// Export writes all identities (cert + matching private key pairs) from
// the named keychain to outFile as an AES-encrypted PKCS#12 bundle, sealed
// with password. Wraps `security export -t identities -f pkcs12`.
func (c *Client) Export(ctx context.Context, keychain, outFile, password string) error {
	keyPath, err := keychainPath(keychain)
	if err != nil {
		return mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"security.Export", "could not resolve keychain path",
			mferrors.WithCause(err))
	}
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "security",
		Args: []string{
			"export", "-k", keyPath,
			"-t", "identities",
			"-f", "pkcs12",
			"-P", password,
			"-o", outFile,
		},
		Redact: []string{password},
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"security.Export",
			fmt.Sprintf("security export failed: %s", strings.TrimSpace(string(res.Stderr))),
			mferrors.WithDetails(map[string]any{"keychain": keychain, "out": outFile, "exit": res.ExitCode}))
	}
	return nil
}

// Import installs an X.509 certificate or PKCS#12 key+cert bundle into a
// keychain. -A allows any application to use the imported item.
func (c *Client) Import(ctx context.Context, file, keychain, password string) error {
	args := []string{"import", file, "-k", keychain, "-A"}
	redact := []string{}
	if password != "" {
		args = append(args, "-P", password)
		redact = append(redact, password)
	}
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool:   "security",
		Args:   args,
		Redact: redact,
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewIdentity(mferrors.CodeIdentityImportFail,
			"security.Import",
			fmt.Sprintf("security import failed: %s", strings.TrimSpace(string(res.Stderr))),
			mferrors.WithDetails(map[string]any{"file": file, "keychain": keychain}))
	}
	return nil
}
