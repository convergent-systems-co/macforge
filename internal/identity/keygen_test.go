// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity_test

import (
	"testing"

	"github.com/convergent-systems-co/macforge/internal/identity"
)

func TestGenerateRSAKey_Is2048Bit(t *testing.T) {
	key, err := identity.GenerateRSAKey()
	if err != nil {
		t.Fatalf("GenerateRSAKey: %v", err)
	}
	if got := key.N.BitLen(); got != 2048 {
		t.Fatalf("key bit length = %d, want 2048 (Apple Developer ID requires RSA-2048)", got)
	}
}
