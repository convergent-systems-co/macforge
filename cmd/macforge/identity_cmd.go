// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/apple/security"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
	"github.com/convergent-systems-co/macforge/internal/identity"
	"github.com/convergent-systems-co/macforge/internal/keychain"
)

func newIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Manage Developer ID identities",
	}
	cmd.AddCommand(
		newIdentityCreateCmd(),
		newIdentityImportCmd(),
		newIdentityListCmd(),
		stubSub("rotate", "Rotate the certificate; archive the old (v0.2)"),
		newIdentityStatusCmd(),
		stubSub("export", "Encrypted export for CI consumption (v0.2)"),
	)
	return cmd
}

func newIdentityCreateCmd() *cobra.Command {
	var (
		cn, email, country string
		org                string
		outPrefix          string
		p12Password        string
		keychainName       string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a new RSA-2048 private key + CSR; bundle key as PKCS#12 + import into the macforge keychain",
		Long: `Generates a fresh RSA-2048 private key, writes a PKCS#10 CSR for the
Apple Developer ID portal, and imports the same private key (with a
self-signed placeholder cert) into the configured macforge keychain.

The private key never touches disk in unencrypted form — the PKCS#12
backup is AES-encrypted with the password from --p12-password, or with a
fresh randomly generated password shown ONCE in the result envelope
(save it immediately — it cannot be recovered).

Upload the resulting .csr to https://developer.apple.com/account/resources/certificates
and pick "Developer ID Application". When Apple returns the issued .cer,
import it with: macforge identity import --file <cert.cer>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("identity.create", true)
			if err != nil {
				return err
			}
			if keychainName == "" {
				keychainName = keychain.DefaultName(rt.cfg.Team, "signing")
			}
			svc := identity.New(security.New(newRunnerWithAudit(rt)))
			result, runErr := svc.Create(cmd.Context(), identity.CreateOptions{
				Subject: identity.CSRSubject{
					CommonName:   cn,
					Organization: org,
					Email:        email,
					Country:      country,
				},
				Keychain:    keychainName,
				OutPrefix:   outPrefix,
				P12Password: p12Password,
			})
			return rt.emit(identityCreateResult{Result: result}, runErr)
		},
	}
	cmd.Flags().StringVar(&cn, "cn", "", "Common Name for the CSR (required)")
	cmd.Flags().StringVar(&org, "org", "", "Organization for the CSR (optional)")
	cmd.Flags().StringVar(&email, "email", "", "Email Address for the CSR (optional)")
	cmd.Flags().StringVar(&country, "country", "US", "ISO 3166 two-letter country code")
	cmd.Flags().StringVar(&outPrefix, "out", "./identity", "output path prefix; <prefix>.csr and <prefix>.p12 are written")
	cmd.Flags().StringVar(&p12Password, "p12-password", "", "password for the .p12 backup (omit to generate a random one)")
	cmd.Flags().StringVar(&keychainName, "keychain", "", "target keychain (default: macforge-<team>-signing)")
	_ = cmd.MarkFlagRequired("cn")
	return cmd
}

// stubSub returns a cobra.Command that always errors with MF-CONFIG-001
// "not yet implemented". Used for subverbs deferred past v0.1.
// Note: this is also referenced from keychain_cmd.go before Task 14's
// replacement; once both files are at their post-Task-15 versions, this
// is the sole definition of stubSub in the package.
func stubSub(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", cmd.CommandPath(), "not yet implemented")
		},
	}
}

func newIdentityImportCmd() *cobra.Command {
	var file, keychainName, p12Password string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import Developer ID certificate(s) into the dedicated keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("identity.import", true)
			if err != nil {
				return err
			}
			if keychainName == "" {
				keychainName = keychain.DefaultName(rt.cfg.Team, "signing")
			}
			svc := identity.New(security.New(newRunnerWithAudit(rt)))
			runErr := svc.Import(cmd.Context(), identity.ImportOptions{
				File:        file,
				Keychain:    keychainName,
				P12Password: p12Password,
			})
			return rt.emit(identityImportResult{File: file, Keychain: keychainName}, runErr)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "path to .cer, .pem, or .p12")
	cmd.Flags().StringVar(&keychainName, "keychain", "", "target keychain (default: macforge-<team>-signing)")
	cmd.Flags().StringVar(&p12Password, "p12-password", "", "password for .p12 file (omit for .cer/.pem)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newIdentityListCmd() *cobra.Command {
	var keychainName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List identities in the configured keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("identity.list", true)
			if err != nil {
				return err
			}
			if keychainName == "" {
				keychainName = keychain.DefaultName(rt.cfg.Team, "signing")
			}
			svc := identity.New(security.New(newRunnerWithAudit(rt)))
			ids, runErr := svc.List(cmd.Context(), keychainName)
			return rt.emit(identityListResult{Keychain: keychainName, Identities: ids}, runErr)
		},
	}
	cmd.Flags().StringVar(&keychainName, "keychain", "", "keychain to query (default: macforge-<team>-signing)")
	return cmd
}

func newIdentityStatusCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show certificate validity, expiration, team (parsed from a cert file)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("identity.status", false)
			if err != nil {
				return err
			}
			st, runErr := identity.ReadCertStatus(file)
			return rt.emit(identityStatusResult{File: file, Status: st}, runErr)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "path to .cer or .pem certificate file")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

// Result types ----------------------------------------------------------------

type identityCreateResult struct {
	Result identity.CreateResult `json:"result"`
}

func (r identityCreateResult) SchemaName() string { return "macforge.v1.identity.create" }
func (r identityCreateResult) HumanLines() []string {
	out := []string{
		"CSR written:          " + r.Result.CSRPath,
		"  (upload at https://developer.apple.com/account/resources/certificates → Developer ID Application)",
		"PKCS#12 key backup:   " + r.Result.P12Path,
	}
	if r.Result.GeneratedP12Password != "" {
		out = append(out,
			"Generated password:   "+r.Result.GeneratedP12Password,
			"  ⚠  SAVE THIS PASSWORD NOW — it is not stored and will not be shown again.",
		)
	} else {
		out = append(out, "Password:             (provided via --p12-password)")
	}
	out = append(out,
		"Keychain import:      "+r.Result.Keychain+"  (the key is bound to a self-signed placeholder",
		"                       cert that will be ignored once you `macforge identity import` the",
		"                       real Apple-issued .cer file)",
		"Public key SHA-256:   "+r.Result.PublicKeyFingerprint[:32]+"…",
	)
	return out
}

type identityImportResult struct {
	File     string `json:"file"`
	Keychain string `json:"keychain"`
}

func (r identityImportResult) SchemaName() string { return "macforge.v1.identity.import" }
func (r identityImportResult) HumanLines() []string {
	return []string{"Imported: " + r.File, "Keychain: " + r.Keychain}
}

type identityListResult struct {
	Keychain   string              `json:"keychain"`
	Identities []security.Identity `json:"identities"`
}

func (r identityListResult) SchemaName() string { return "macforge.v1.identity.list" }
func (r identityListResult) HumanLines() []string {
	out := []string{"Keychain: " + r.Keychain}
	if len(r.Identities) == 0 {
		return append(out, "  (no identities found)")
	}
	for _, id := range r.Identities {
		out = append(out, "  "+id.SHA1Fingerprint+"  "+id.CommonName)
	}
	return out
}

type identityStatusResult struct {
	File   string               `json:"file"`
	Status identity.CertStatus  `json:"status"`
}

func (r identityStatusResult) SchemaName() string { return "macforge.v1.identity.status" }
func (r identityStatusResult) HumanLines() []string {
	exp := "valid"
	if r.Status.Expired {
		exp = "EXPIRED"
	}
	return []string{
		"File:           " + r.File,
		"Subject:        " + r.Status.Subject,
		"Issuer:         " + r.Status.Issuer,
		"NotBefore:      " + r.Status.NotBefore.Format("2006-01-02"),
		"NotAfter:       " + r.Status.NotAfter.Format("2006-01-02"),
		"Days to expiry: " + intToStr(r.Status.DaysToExpiry),
		"Status:         " + exp,
	}
}

func intToStr(i int) string {
	// avoid importing strconv just for this one site
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		buf[n] = '-'
	}
	return string(buf[n:])
}
