package config

import (
	"strings"
	"testing"
)

func TestRuntime_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		verbose bool
		quiet   bool
		wantErr string
	}{
		{name: "both off", verbose: false, quiet: false, wantErr: ""},
		{name: "only verbose", verbose: true, quiet: false, wantErr: ""},
		{name: "only quiet", verbose: false, quiet: true, wantErr: ""},
		{name: "both on", verbose: true, quiet: true, wantErr: "mutually exclusive"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rt := &Runtime{Verbose: tc.verbose, Quiet: tc.quiet}
			err := rt.Validate()
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("want nil error, got %v", err)
			case tc.wantErr != "" && err == nil:
				t.Fatalf("want error containing %q, got nil", tc.wantErr)
			case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
				t.Fatalf("want error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestRuntime_VersionString(t *testing.T) {
	t.Parallel()

	rt := &Runtime{
		Version:   "v0.0.1",
		Commit:    "deadbeef",
		BuildDate: "2026-01-01T00:00:00Z",
	}
	got := rt.VersionString()
	want := "v0.0.1 (commit deadbeef, built 2026-01-01T00:00:00Z)"
	if got != want {
		t.Fatalf("VersionString:\n  got:  %q\n  want: %q", got, want)
	}
}
