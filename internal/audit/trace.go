// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// NewTraceID returns a fresh ULID for a single macforge invocation.
// The same trace threads through every event in that invocation.
func NewTraceID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader).String()
}
