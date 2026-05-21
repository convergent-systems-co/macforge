// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// FakeRunner replays recorded Apple-tool invocations. Used by Tier 2
// integration tests; never used in production.
type FakeRunner struct {
	fixtures []fixture
}

type fixture struct {
	tool   string
	args   []string
	stdout []byte
	stderr []byte
	exit   int
}

// NewFakeRunner loads every fixture from root/<tool>/<scenario>/. The
// inv.json file identifies the invocation; stdout, stderr, and exit
// hold the replay payload.
func NewFakeRunner(root string) (*FakeRunner, error) {
	r := &FakeRunner{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "inv.json" {
			return nil
		}
		dir := filepath.Dir(path)

		invBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var spec struct {
			Tool string   `json:"tool"`
			Args []string `json:"args"`
		}
		if err := json.Unmarshal(invBytes, &spec); err != nil {
			return fmt.Errorf("fixture %s: %w", path, err)
		}

		stdout, _ := os.ReadFile(filepath.Join(dir, "stdout"))
		stderr, _ := os.ReadFile(filepath.Join(dir, "stderr"))
		exitStr, _ := os.ReadFile(filepath.Join(dir, "exit"))
		exit, _ := strconv.Atoi(strings.TrimSpace(string(exitStr)))

		// Normalize CRLF → LF on read. Fixture files SHOULD be LF-only via
		// .gitattributes, but defending against git autocrlf misconfiguration
		// on Windows clones keeps tests platform-portable.
		r.fixtures = append(r.fixtures, fixture{
			tool:   spec.Tool,
			args:   spec.Args,
			stdout: normalizeLF(stdout),
			stderr: normalizeLF(stderr),
			exit:   exit,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// normalizeLF converts CRLF to LF in fixture bodies so tests behave the
// same whether the fixture was checked out on a Unix-y system or a Windows
// box with autocrlf enabled.
func normalizeLF(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n' {
			continue
		}
		out = append(out, b[i])
	}
	return out
}

// Run searches loaded fixtures for an exact (tool, args) match.
func (r *FakeRunner) Run(ctx context.Context, inv Invocation) (Result, error) {
	for _, f := range r.fixtures {
		if f.tool == inv.Tool && reflect.DeepEqual(f.args, inv.Args) {
			return Result{
				ExitCode: f.exit,
				Stdout:   f.stdout,
				Stderr:   f.stderr,
			}, nil
		}
	}
	return Result{}, fmt.Errorf("FakeRunner: no fixture matches tool=%q args=%v (have %d fixtures)",
		inv.Tool, inv.Args, len(r.fixtures))
}
