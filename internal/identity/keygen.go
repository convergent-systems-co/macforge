// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package identity

import (
	"crypto/rand"
	"crypto/rsa"
)

// GenerateRSAKey returns a new RSA-2048 private key. Apple Developer ID
// Application certificates require RSA-2048; EC P-256 is not accepted by
// the Developer ID issuer.
func GenerateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}
