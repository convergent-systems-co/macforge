// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package audit writes the MacForge JSONL audit log. The schema mirrors the
// vocabulary defined in ~/.ai/Common.md §5.2 so the same grep patterns work
// across MacForge logs and the user's broader governance audit streams.
//
// See docs/adr/0012-audit-log-schema.md.
package audit
