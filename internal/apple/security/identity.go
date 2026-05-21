// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package security

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/apple"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// Identity describes one signing identity found in a keychain.
type Identity struct {
	SHA1Fingerprint string // 40 hex chars
	CommonName      string // e.g. "Developer ID Application: Org (TEAM)"
}

// FindIdentities lists valid identities matching the given policy
// ("codesigning" for Developer ID Application).
func (c *Client) FindIdentities(ctx context.Context, keychain, policy string) ([]Identity, error) {
	res, err := c.r.Run(ctx, apple.Invocation{
		Tool: "security",
		Args: argsFindIdentity(keychain, policy),
	})
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, mferrors.NewIdentity(mferrors.CodeIdentityNotFound,
			"security.FindIdentities",
			fmt.Sprintf("security find-identity failed: %s", strings.TrimSpace(string(res.Stderr))))
	}
	return parseFindIdentity(string(res.Stdout)), nil
}

// parseFindIdentity extracts identities from `security find-identity` output:
//
//	1) ABCDEF0123...  "Developer ID Application: Org Inc. (XYZ1234567)"
//	     ... matching identities found
//	     ... valid identities found
var reFindIdentityLine = regexp.MustCompile(`^\s*\d+\)\s+([0-9A-Fa-f]{40})\s+"([^"]+)"`)

func parseFindIdentity(stdout string) []Identity {
	var out []Identity
	for _, line := range strings.Split(stdout, "\n") {
		m := reFindIdentityLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		out = append(out, Identity{SHA1Fingerprint: m[1], CommonName: m[2]})
	}
	return out
}
