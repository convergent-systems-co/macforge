package cmd

import (
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestStatus_RunsAndProducesOutput exercises cmd -> status.Run wiring.
// status.Run never returns an ExitCoder (status is informational), but we
// still swap cli.OsExiter as belt-and-suspenders in case a future section
// decides to fail — mirrors cmd/doctor_test.go's pattern.
func TestStatus_RunsAndProducesOutput(t *testing.T) {
	var captured int
	prevExiter := cli.OsExiter
	cli.OsExiter = func(code int) { captured = code }
	t.Cleanup(func() { cli.OsExiter = prevExiter })

	stdout, _, _, err := runRoot(t, "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != 0 {
		t.Errorf("status should not call OsExiter; got captured=%d", captured)
	}
	for _, want := range []string{"brew", "repo"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q\nfull:\n%s", want, stdout)
		}
	}
}
