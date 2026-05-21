// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package apple is the one boundary between MacForge and Apple's CLI tools.
// Every shell-out flows through Runner; per-tool wrappers under apple/<tool>/
// build Invocations and parse Results. See ADR-0003.
package apple

import "context"

// Runner abstracts the execution of an Apple-tool invocation. The two
// implementations are ExecRunner (real os/exec) and FakeRunner (fixture
// replay). All other code depends on this interface only.
type Runner interface {
	Run(ctx context.Context, inv Invocation) (Result, error)
}
