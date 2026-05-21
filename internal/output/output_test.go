// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
	"github.com/convergent-systems-co/macforge/internal/output"
)

type fakeResult struct {
	Artifact string `json:"artifact"`
	Team     string `json:"team"`
}

func (f fakeResult) SchemaName() string  { return "macforge.v1.fake" }
func (f fakeResult) HumanLines() []string { return []string{"Artifact: " + f.Artifact, "Team:     " + f.Team} }

func TestRender_JSON_Success(t *testing.T) {
	var buf bytes.Buffer
	r := output.NewRenderer(output.ModeJSON, &buf, "01HV", "sign", false)
	if err := r.Success(fakeResult{Artifact: "./MyApp.app", Team: "XYZ1234567"}); err != nil {
		t.Fatalf("Success: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if got["ok"] != true {
		t.Fatalf("ok = %v, want true", got["ok"])
	}
	if got["schema"] != "macforge.v1.fake" {
		t.Fatalf("schema = %v, want macforge.v1.fake", got["schema"])
	}
	if got["command"] != "sign" {
		t.Fatalf("command = %v, want sign", got["command"])
	}
	if got["trace"] != "01HV" {
		t.Fatalf("trace = %v, want 01HV", got["trace"])
	}
}

func TestRender_Human_Success(t *testing.T) {
	var buf bytes.Buffer
	r := output.NewRenderer(output.ModeHuman, &buf, "01HV", "sign", true /* nocolor */)
	if err := r.Success(fakeResult{Artifact: "./MyApp.app", Team: "XYZ"}); err != nil {
		t.Fatalf("Success: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Artifact: ./MyApp.app") {
		t.Fatalf("human output missing artifact line:\n%s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("human output contains ANSI when nocolor=true:\n%q", out)
	}
}

func TestRender_JSON_Error(t *testing.T) {
	var buf bytes.Buffer
	r := output.NewRenderer(output.ModeJSON, &buf, "01HV", "notarize", false)
	mfErr := mferrors.NewNotarize(mferrors.CodeNotarizeRejected, "notarize.Submit", "rejected by Apple",
		mferrors.WithHint("Check identity status"))

	if err := r.Failure(mfErr); err != nil {
		t.Fatalf("Failure: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if got["ok"] != false {
		t.Fatalf("ok = %v, want false", got["ok"])
	}
	if got["code"] != "MF-NOTARIZE-001" {
		t.Fatalf("code = %v, want MF-NOTARIZE-001", got["code"])
	}
	if got["op"] != "notarize.Submit" {
		t.Fatalf("op = %v, want notarize.Submit", got["op"])
	}
	if got["hint"] != "Check identity status" {
		t.Fatalf("hint = %v, want 'Check identity status'", got["hint"])
	}
}
