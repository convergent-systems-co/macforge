// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/audit"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
	"github.com/convergent-systems-co/macforge/internal/keychain"
)

func newKeychainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keychain",
		Short: "Manage dedicated MacForge keychains",
	}
	cmd.AddCommand(
		newKeychainCreateCmd(),
		newKeychainDeleteCmd(),
		newKeychainListCmd(),
		newKeychainUnlockCmd(),
	)
	return cmd
}

// runner builds the ExecRunner with audit wiring for this invocation.
func newRunnerWithAudit(rt *cliRuntime) *apple.ExecRunner {
	r := apple.NewExecRunner(rt.audit)
	// Trace wiring — ExecRunner stamps the trace onto every audit event.
	type traceSetter interface{ SetTrace(string) }
	if ts, ok := any(r).(traceSetter); ok {
		ts.SetTrace(rt.trace)
	}
	_ = audit.ActorMacforge // keep audit imported even if helpers shift
	return r
}

func newKeychainCreateCmd() *cobra.Command {
	var name, secretRef string
	var allowNonstandard bool
	var lockOnSleep bool
	var lockTimeout int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new managed keychain with the default name",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("apple.keychain.create", true)
			if err != nil {
				return err
			}

			result, runErr := func() (keychainCreateResult, error) {
				if name == "" {
					name = keychain.DefaultName(rt.cfg.Team, "signing")
				}
				m := keychain.NewManager(security.New(newRunnerWithAudit(rt)))
				err := m.Create(cmd.Context(), keychain.CreateOptions{
					Name:             name,
					SecretRef:        secretRef,
					AllowNonstandard: allowNonstandard,
					LockOnSleep:      lockOnSleep,
					LockTimeoutSecs:  lockTimeout,
				})
				if err != nil {
					return keychainCreateResult{}, err
				}
				return keychainCreateResult{Name: name, Team: rt.cfg.Team}, nil
			}()
			return rt.emit(result, runErr)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "keychain name (default: macforge-<team>-signing)")
	cmd.Flags().StringVar(&secretRef, "secret-ref", "env:MACFORGE_KEYCHAIN_PASSWORD", "env:VAR or keyring:macforge:<id>")
	cmd.Flags().BoolVar(&allowNonstandard, "allow-nonstandard", false, "skip the macforge-<TEAM>-<PURPOSE> naming validator")
	cmd.Flags().BoolVar(&lockOnSleep, "lock-on-sleep", true, "lock keychain when the system sleeps")
	cmd.Flags().IntVar(&lockTimeout, "lock-timeout", 3600, "auto-lock after N seconds; 0 disables")
	_ = mferrors.ErrKeychain
	return cmd
}

func newKeychainDeleteCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a managed keychain (refuses login.keychain)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("apple.keychain.delete", true)
			if err != nil {
				return err
			}
			if name == "" {
				name = keychain.DefaultName(rt.cfg.Team, "signing")
			}
			m := keychain.NewManager(security.New(newRunnerWithAudit(rt)))
			runErr := m.Delete(cmd.Context(), name)
			return rt.emit(keychainDeleteResult{Name: name}, runErr)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "keychain name (default: macforge-<team>-signing)")
	return cmd
}

func newKeychainListCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List signing identities in a managed keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("apple.keychain.list", true)
			if err != nil {
				return err
			}
			if name == "" {
				name = keychain.DefaultName(rt.cfg.Team, "signing")
			}
			m := keychain.NewManager(security.New(newRunnerWithAudit(rt)))
			ids, runErr := m.List(cmd.Context(), name)
			return rt.emit(keychainListResult{Keychain: name, Identities: ids}, runErr)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "keychain name (default: macforge-<team>-signing)")
	return cmd
}

func newKeychainUnlockCmd() *cobra.Command {
	var name, secretRef string
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock the configured keychain for the current shell session",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("apple.keychain.unlock", true)
			if err != nil {
				return err
			}
			if name == "" {
				name = keychain.DefaultName(rt.cfg.Team, "signing")
			}
			m := keychain.NewManager(security.New(newRunnerWithAudit(rt)))
			runErr := m.Unlock(cmd.Context(), name, secretRef)
			return rt.emit(keychainUnlockResult{Name: name}, runErr)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "keychain name (default: macforge-<team>-signing)")
	cmd.Flags().StringVar(&secretRef, "secret-ref", "env:MACFORGE_KEYCHAIN_PASSWORD", "env:VAR or keyring:macforge:<id>")
	return cmd
}

// Result types (Outputter implementations) ------------------------------------

type keychainCreateResult struct {
	Name string `json:"name"`
	Team string `json:"team"`
}

func (r keychainCreateResult) SchemaName() string   { return "macforge.v1.apple.keychain.create" }
func (r keychainCreateResult) HumanLines() []string { return []string{"Created: " + r.Name, "Team:    " + r.Team} }

type keychainDeleteResult struct {
	Name string `json:"name"`
}

func (r keychainDeleteResult) SchemaName() string   { return "macforge.v1.apple.keychain.delete" }
func (r keychainDeleteResult) HumanLines() []string { return []string{"Deleted: " + r.Name} }

type keychainListResult struct {
	Keychain   string              `json:"keychain"`
	Identities []security.Identity `json:"identities"`
}

func (r keychainListResult) SchemaName() string { return "macforge.v1.apple.keychain.list" }
func (r keychainListResult) HumanLines() []string {
	out := []string{"Keychain: " + r.Keychain}
	if len(r.Identities) == 0 {
		out = append(out, "  (no identities found)")
		return out
	}
	for _, id := range r.Identities {
		out = append(out, "  "+id.SHA1Fingerprint+"  "+id.CommonName)
	}
	return out
}

type keychainUnlockResult struct {
	Name string `json:"name"`
}

func (r keychainUnlockResult) SchemaName() string   { return "macforge.v1.apple.keychain.unlock" }
func (r keychainUnlockResult) HumanLines() []string { return []string{"Unlocked: " + r.Name} }
