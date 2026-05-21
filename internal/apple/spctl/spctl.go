// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package spctl wraps the macOS spctl CLI for Gatekeeper assessment.
package spctl

import (
	"context"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/apple"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// AssessType is the spctl --type value.
type AssessType string

const (
	// AssessTypeExec assesses a .app bundle (Gatekeeper "execute" policy).
	AssessTypeExec AssessType = "execute"
	// AssessTypeInstall assesses a .pkg installer ("install" policy).
	AssessTypeInstall AssessType = "install"
	// AssessTypeOpen assesses a .dmg disk image ("open" policy).
	AssessTypeOpen AssessType = "open"
)

// Client wraps spctl via a Runner.
type Client struct {
	r apple.Runner
}

// New returns a Client driven by r.
func New(r apple.Runner) *Client { return &Client{r: r} }

// Result is the typed outcome of an Assess call.
type Result struct {
	Path     string `json:"path"`
	Accepted bool   `json:"accepted"`
	Source   string `json:"source"` // e.g. "Notarized Developer ID"
	Stderr   string `json:"stderr"`
}

// Assess runs spctl --assess for the given path and type.
func (c *Client) Assess(ctx context.Context, path string, t AssessType) (Result, error) {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "spctl",
		Args: argsAssess(path, t),
	})
	if err != nil {
		return Result{}, err
	}
	r := Result{Path: path, Stderr: string(res.Stderr), Accepted: res.ExitCode == 0}
	r.Source = parseSource(string(res.Stderr))
	if !r.Accepted {
		return r, mferrors.NewVerify(mferrors.CodeVerifySpctlFail,
			"spctl.Assess",
			"spctl rejected "+path+": "+strings.TrimSpace(string(res.Stderr)),
			mferrors.WithDetails(map[string]any{"path": path, "type": string(t)}))
	}
	return r, nil
}

func argsAssess(path string, t AssessType) []string {
	return []string{"--assess", "--type", string(t), "--verbose=4", path}
}

func parseSource(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "source=") {
			return strings.TrimPrefix(line, "source=")
		}
	}
	return ""
}
