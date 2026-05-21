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
