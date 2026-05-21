// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config_test

import (
	"testing"

	"github.com/convergent-systems-co/macforge/internal/config"
)

func TestResolveKeychainName(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "name unset → derive from team",
			cfg: &config.Config{
				Version: 1,
				Team:    "XYZ1234567",
			},
			want: "macforge-XYZ1234567-signing",
		},
		{
			name: "name set + canonical → return verbatim",
			cfg: &config.Config{
				Version: 1,
				Team:    "XYZ1234567",
				Keychain: config.KeychainConfig{
					Name: "macforge-XYZ1234567-signing",
				},
			},
			want: "macforge-XYZ1234567-signing",
		},
		{
			name: "name set + nonstandard + allow=true → return verbatim",
			cfg: &config.Config{
				Version: 1,
				Team:    "XYZ1234567",
				Keychain: config.KeychainConfig{
					Name:             "my-weird-but-allowed-name",
					AllowNonstandard: true,
				},
			},
			want: "my-weird-but-allowed-name",
		},
		{
			name: "name set + non-default purpose → return verbatim",
			cfg: &config.Config{
				Version: 1,
				Team:    "XYZ1234567",
				Keychain: config.KeychainConfig{
					Name: "macforge-XYZ1234567-custom",
				},
			},
			want: "macforge-XYZ1234567-custom",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := config.ResolveKeychainName(tc.cfg)
			if got != tc.want {
				t.Fatalf("ResolveKeychainName = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolveKeychainName_NilConfig(t *testing.T) {
	if got := config.ResolveKeychainName(nil); got != "" {
		t.Fatalf("ResolveKeychainName(nil) = %q, want empty string", got)
	}
}
