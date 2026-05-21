// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func sha256Hex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func osGetwd() (string, error) { return os.Getwd() }
