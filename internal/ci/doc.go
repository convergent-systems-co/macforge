// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package ci detects the host CI provider (GitHub Actions, GitLab, Azure DevOps,
// or none) and exposes provider-specific helpers (workflow-command emission,
// secret reading, artifact upload paths).
package ci
