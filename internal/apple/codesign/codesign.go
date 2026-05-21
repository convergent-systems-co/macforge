// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package codesign wraps the macOS codesign CLI for signing and verification.
package codesign

import (
	"context"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/apple"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// Client drives codesign via a Runner.
type Client struct {
	r apple.Runner
}

// New returns a Client bound to r.
func New(r apple.Runner) *Client { return &Client{r: r} }

// SignOptions describes one sign call.
type SignOptions struct {
	IdentityCN      string // "Developer ID Application: ..."
	Keychain        string // dedicated keychain name
	HardenedRuntime bool
	Timestamp       bool
	Entitlements    string // path to plist; empty to skip
	Path            string // artifact to sign
}

// Sign runs codesign --sign with the configured options.
func (c *Client) Sign(ctx context.Context, opts SignOptions) error {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "codesign",
		Args: argsSign(opts),
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewSigning(mferrors.CodeSignVerificationFail,
			"codesign.Sign",
			"codesign --sign failed: "+strings.TrimSpace(string(res.Stderr)),
			mferrors.WithDetails(map[string]any{"path": opts.Path, "identity": opts.IdentityCN}))
	}
	return nil
}

// Verify runs codesign --verify --strict --verbose=2.
func (c *Client) Verify(ctx context.Context, path string) (VerifyResult, error) {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "codesign",
		Args: argsVerify(path),
	})
	if err != nil {
		return VerifyResult{}, err
	}
	if res.ExitCode != 0 {
		return VerifyResult{Path: path, OK: false, Stderr: string(res.Stderr)},
			mferrors.NewVerify(mferrors.CodeVerifyCodesignFail,
				"codesign.Verify",
				strings.TrimSpace(string(res.Stderr)),
				mferrors.WithDetails(map[string]any{"path": path}))
	}
	return VerifyResult{Path: path, OK: true, Stderr: string(res.Stderr)}, nil
}

// Display runs codesign --display --verbose=4 and returns parsed identity info.
func (c *Client) Display(ctx context.Context, path string) (DisplayInfo, error) {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "codesign",
		Args: argsDisplay(path),
	})
	if err != nil {
		return DisplayInfo{}, err
	}
	if res.ExitCode != 0 {
		return DisplayInfo{}, mferrors.NewVerify(mferrors.CodeVerifyCodesignFail,
			"codesign.Display",
			strings.TrimSpace(string(res.Stderr)))
	}
	return parseDisplay(string(res.Stderr)), nil
}

// VerifyResult is the typed outcome of Verify.
type VerifyResult struct {
	Path   string `json:"path"`
	OK     bool   `json:"ok"`
	Stderr string `json:"stderr"`
}

// DisplayInfo is parsed from codesign --display output.
type DisplayInfo struct {
	Identifier string `json:"identifier"`
	TeamID     string `json:"team_id"`
	Authority  string `json:"authority"`
	Timestamp  string `json:"timestamp"`
}

func argsSign(o SignOptions) []string {
	args := []string{"--sign", o.IdentityCN}
	if o.Keychain != "" {
		args = append(args, "--keychain", o.Keychain)
	}
	if o.HardenedRuntime {
		args = append(args, "--options", "runtime")
	}
	if o.Timestamp {
		args = append(args, "--timestamp")
	}
	if o.Entitlements != "" {
		args = append(args, "--entitlements", o.Entitlements)
	}
	args = append(args, "--verbose")
	return append(args, o.Path)
}

func argsVerify(path string) []string {
	return []string{"--verify", "--strict", "--verbose=2", path}
}

func argsDisplay(path string) []string {
	return []string{"--display", "--verbose=4", path}
}
