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

		r.fixtures = append(r.fixtures, fixture{
			tool:   spec.Tool,
			args:   spec.Args,
			stdout: stdout,
			stderr: stderr,
			exit:   exit,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
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
