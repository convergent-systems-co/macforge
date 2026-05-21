// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/config"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// newConfigCmd returns the `apple config` subtree. Today it holds one
// verb (`validate`); future config-shaped verbs (`show`, `migrate`, ...)
// land alongside it. See ADR-0019 for the validation contract.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Apple config inspection (validate)",
	}
	cmd.AddCommand(newConfigValidateCmd())
	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Run static and runtime checks against the loaded MacForge config",
		Long: `Validate the MacForge config layered from defaults + global + project + env.

Static checks (run as part of config.Load and re-summarized here):
  - schema version
  - top-level team is non-empty
  - keychain.name matches macforge-<TEAM>-<PURPOSE> when set
  - keychain.name's <TEAM> segment equals top-level team (unless allow_nonstandard: true)
  - keychain.unlock is env:VAR or keyring:VAR (never an inline secret)

Runtime checks (extra; performed only by this verb):
  - env-var referenced by keychain.unlock is set in the current process
  - keychain file referenced by name is reachable via security show-keychain-info

Exit code is non-zero if any check is red.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("apple.config.validate", true)
			if err != nil {
				// config.Load itself failed (structural error). Surface it
				// through the same emit path as every other verb so the
				// renderer prints a proper ✗ envelope.
				return rt.emit(configValidateResult{}, err)
			}

			result, runErr := runConfigValidate(cmd, rt)
			return rt.emit(result, runErr)
		},
	}
}

// runConfigValidate walks rt.cfg field-by-field and produces a check list.
// Static checks here are a re-summary of what config.Load already enforced
// (so a green run means "Load accepted this config"); runtime checks add
// reachability tests for env vars and the keychain file on disk.
func runConfigValidate(cmd *cobra.Command, rt *cliRuntime) (configValidateResult, error) {
	cfg := rt.cfg
	res := configValidateResult{}

	// ---- Static recap (everything below already passed config.Load, but
	//      we surface each rule individually so operators see WHAT validated). --

	res.add(checkOK(fmt.Sprintf("config schema version = %d", cfg.Version)))
	res.add(checkOK(fmt.Sprintf("team = %s", cfg.Team)))

	// Keychain name resolution: show what the resolver returns. If the user
	// set keychain.name, also call it out separately so they can see the
	// resolution chain.
	resolved := config.ResolveKeychainName(cfg)
	if cfg.Keychain.Name != "" {
		if cfg.Keychain.AllowNonstandard {
			res.add(checkOK(fmt.Sprintf("keychain.name = %s (nonstandard, opted in)", cfg.Keychain.Name)))
		} else {
			res.add(checkOK(fmt.Sprintf("keychain.name = %s (canonical, matches team %s)", cfg.Keychain.Name, cfg.Team)))
		}
	} else {
		res.add(checkOK(fmt.Sprintf("keychain.name = %s (derived from team)", resolved)))
	}

	// Unlock reference format.
	switch {
	case cfg.Keychain.Unlock == "":
		res.add(checkInfo("keychain.unlock = (unset; required for non-interactive signing)"))
	case strings.HasPrefix(cfg.Keychain.Unlock, "env:"):
		res.add(checkOK(fmt.Sprintf("keychain.unlock = %s", cfg.Keychain.Unlock)))
	case strings.HasPrefix(cfg.Keychain.Unlock, "keyring:"):
		res.add(checkOK(fmt.Sprintf("keychain.unlock = %s", cfg.Keychain.Unlock)))
	default:
		// Load would have already rejected this; defensive.
		res.add(checkFail(
			fmt.Sprintf("keychain.unlock = %q is neither env: nor keyring:", cfg.Keychain.Unlock),
			"set keychain.unlock: env:VAR_NAME with the env var holding the keychain password",
		))
	}

	// ---- Runtime: env var presence (don't read its value, just check it's set). --

	if strings.HasPrefix(cfg.Keychain.Unlock, "env:") {
		varName := strings.TrimPrefix(cfg.Keychain.Unlock, "env:")
		if _, present := os.LookupEnv(varName); present {
			res.add(checkOK(fmt.Sprintf("$%s is set", varName)))
		} else {
			res.add(checkFail(
				fmt.Sprintf("$%s is unset", varName),
				fmt.Sprintf("export %s=\"<your-keychain-password>\"", varName),
			))
		}
	} else if strings.HasPrefix(cfg.Keychain.Unlock, "keyring:") {
		// Keyring lookup isn't wired into this verb yet — surface as info,
		// not red. Follow-up: integrate with internal/keychain.Secret.
		res.add(checkInfo(fmt.Sprintf("%s (keyring reachability not checked this pass)", cfg.Keychain.Unlock)))
	}

	// ---- Runtime: keychain file reachability via security show-keychain-info. --

	ctx := cmd.Context()
	runner := newRunnerWithAudit(rt)
	secClient := security.New(runner)
	present, hkErr := secClient.HasKeychain(ctx, resolved)
	switch {
	case hkErr != nil:
		// security binary itself failed — surface as red.
		res.add(checkFail(
			fmt.Sprintf("could not probe keychain %s (security CLI error)", resolved),
			"check that the security binary is on PATH and that you're running on macOS",
		))
	case present:
		res.add(checkOK(fmt.Sprintf("keychain %s reachable", resolved)))
	default:
		res.add(checkFail(
			fmt.Sprintf("keychain %s not found on disk", resolved),
			"run `macforge apple keychain create` to create it",
		))
	}

	// ---- Static informational fields (○) — visible but not validated. -----------

	if cfg.Sign.HardenedRuntime {
		res.add(checkOK("sign.hardened_runtime = true"))
	} else {
		res.add(checkInfo("sign.hardened_runtime = false (allowed but unusual for Developer ID)"))
	}
	if cfg.Sign.Timestamp {
		res.add(checkOK("sign.timestamp = true"))
	} else {
		res.add(checkInfo("sign.timestamp = false (allowed but unusual for Developer ID)"))
	}
	if cfg.Sign.Entitlements == "" {
		res.add(checkInfo("sign.entitlements = (unset; project-shaped, set in ./macforge.yaml)"))
	} else {
		res.add(checkInfo(fmt.Sprintf("sign.entitlements = %s (presence on disk not checked this pass)", cfg.Sign.Entitlements)))
	}
	if cfg.Notarize.ASCProfile != "" {
		res.add(checkInfo(fmt.Sprintf("notarize.asc_profile = %s (notarytool profile presence not checked this pass)", cfg.Notarize.ASCProfile)))
	}

	res.Errors = res.countByStatus(checkStatusFail)
	res.Warnings = res.countByStatus(checkStatusWarn)
	if res.Errors > 0 {
		return res, mferrors.NewConfig(mferrors.CodeConfigInvalid, "apple.config.validate",
			fmt.Sprintf("%d config check(s) failed", res.Errors),
			mferrors.WithHint("Fix the ✗ lines above and re-run `macforge apple config validate`"))
	}
	return res, nil
}

// ---- Result + check types --------------------------------------------------

type checkStatus string

const (
	checkStatusOK   checkStatus = "ok"
	checkStatusFail checkStatus = "fail"
	checkStatusInfo checkStatus = "info"
	checkStatusWarn checkStatus = "warn"
)

type validateCheck struct {
	Status checkStatus `json:"status"`
	Label  string      `json:"label"`
	Hint   string      `json:"hint,omitempty"`
}

func checkOK(label string) validateCheck   { return validateCheck{Status: checkStatusOK, Label: label} }
func checkInfo(label string) validateCheck { return validateCheck{Status: checkStatusInfo, Label: label} }
func checkFail(label, hint string) validateCheck {
	return validateCheck{Status: checkStatusFail, Label: label, Hint: hint}
}

type configValidateResult struct {
	Checks   []validateCheck `json:"checks"`
	Errors   int             `json:"errors"`
	Warnings int             `json:"warnings"`
}

func (r *configValidateResult) add(c validateCheck) { r.Checks = append(r.Checks, c) }

func (r configValidateResult) countByStatus(s checkStatus) int {
	n := 0
	for _, c := range r.Checks {
		if c.Status == s {
			n++
		}
	}
	return n
}

// SchemaName / HumanLines implement output.Outputter.

func (r configValidateResult) SchemaName() string { return "macforge.v1.apple.config.validate" }

func (r configValidateResult) HumanLines() []string {
	out := make([]string, 0, len(r.Checks)+1)
	for _, c := range r.Checks {
		marker := "✓"
		switch c.Status {
		case checkStatusFail:
			marker = "✗"
		case checkStatusInfo:
			marker = "○"
		case checkStatusWarn:
			marker = "!"
		}
		out = append(out, fmt.Sprintf("%s %s", marker, c.Label))
		if c.Hint != "" {
			out = append(out, "  hint: "+c.Hint)
		}
	}
	// Trailing summary line.
	out = append(out, "")
	out = append(out, fmt.Sprintf("%d errors, %d warnings", r.Errors, r.Warnings))
	return out
}
