// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package security

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/apple"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// AddToSearchList prepends keychain to the user's keychain search domain so
// it has lookup priority for find-identity and codesign's automatic identity
// resolution. No-op if already present. Required after `security create-keychain`,
// which creates the file but does NOT update the search list.
func (c *Client) AddToSearchList(ctx context.Context, keychain string) error {
	path, err := keychainPath(keychain)
	if err != nil {
		return mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.AddToSearchList",
			"could not resolve keychain path",
			mferrors.WithCause(err))
	}
	existing, err := c.readSearchList(ctx)
	if err != nil {
		return err
	}
	for _, p := range existing {
		if p == path {
			return nil // already on the search list
		}
	}
	newList := append([]string{path}, existing...)
	return c.writeSearchList(ctx, newList)
}

// RemoveFromSearchList removes keychain from the user's search domain.
// No-op if not present. Call before deleting the keychain file.
func (c *Client) RemoveFromSearchList(ctx context.Context, keychain string) error {
	path, err := keychainPath(keychain)
	if err != nil {
		return mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.RemoveFromSearchList",
			"could not resolve keychain path",
			mferrors.WithCause(err))
	}
	existing, err := c.readSearchList(ctx)
	if err != nil {
		return err
	}
	filtered := make([]string, 0, len(existing))
	for _, p := range existing {
		if p != path {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == len(existing) {
		return nil // not present
	}
	return c.writeSearchList(ctx, filtered)
}

// readSearchList returns the current user-domain search list, one absolute
// path per entry. Empty string entries are dropped.
func (c *Client) readSearchList(ctx context.Context) ([]string, error) {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "security",
		Args: argsListKeychainsRead(),
	})
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.readSearchList",
			strings.TrimSpace(string(res.Stderr)))
	}
	return parseSearchList(string(res.Stdout)), nil
}

// writeSearchList sets the user-domain search list to paths.
func (c *Client) writeSearchList(ctx context.Context, paths []string) error {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "security",
		Args: argsListKeychainsWrite(paths),
	})
	if err != nil {
		return err
	}
	if res.ExitCode != 0 {
		return mferrors.NewKeychain(mferrors.CodeKeychainMissing,
			"security.writeSearchList",
			strings.TrimSpace(string(res.Stderr)),
			mferrors.WithDetails(map[string]any{"paths": paths}))
	}
	return nil
}

// argsListKeychainsRead builds: security list-keychains -d user
func argsListKeychainsRead() []string {
	return []string{"list-keychains", "-d", "user"}
}

// argsListKeychainsWrite builds: security list-keychains -d user -s <paths...>
func argsListKeychainsWrite(paths []string) []string {
	out := append([]string{"list-keychains", "-d", "user", "-s"}, paths...)
	return out
}

// parseSearchList extracts absolute paths from `security list-keychains`
// output. Each line is leading-whitespace + quoted path + newline.
func parseSearchList(stdout string) []string {
	var out []string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, `"`)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// keychainPath resolves the on-disk path of a keychain by name.
// `security create-keychain` writes to ~/Library/Keychains/<name>.keychain-db;
// list-keychains reports that absolute path. If the caller already includes
// the .keychain-db suffix, don't double it.
func keychainPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(name, ".keychain-db") {
		name = name + ".keychain-db"
	}
	return filepath.Join(home, "Library", "Keychains", name), nil
}
