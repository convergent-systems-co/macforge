package cmd

import (
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestDoctor_RunsAndProducesSummary exercises the cmd → doctor.Run wiring.
// macheim doctor runs against the real environment, so its pass/fail mix
// depends on the host. We only assert: (a) the output contains a summary
// line ("checks passed" or "checks failed"); (b) the captured exit code
// is either 0 (all pass) or 1 (any fail); the cli framework's HandleExitCoder
// path is short-circuited by swapping cli.OsExiter so the test process is
// not actually terminated.
func TestDoctor_RunsAndProducesSummary(t *testing.T) {
	var captured int
	prevExiter := cli.OsExiter
	cli.OsExiter = func(code int) { captured = code }
	t.Cleanup(func() { cli.OsExiter = prevExiter })

	stdout, _, _, err := runRoot(t, "doctor")

	// runRoot may have observed Run returning a cli.ExitCoder; in that case
	// OsExiter was invoked and captured holds the exit code. The err value
	// itself is the ExitCoder we expect.
	if err != nil {
		var coder cli.ExitCoder
		if asErrCoder(err, &coder) {
			if coder.ExitCode() != 1 {
				t.Errorf("ExitCoder code: got %d, want 1", coder.ExitCode())
			}
			if captured != 0 && captured != 1 {
				t.Errorf("captured OsExiter code: got %d, want 0 or 1", captured)
			}
		} else {
			t.Fatalf("unexpected error type %T: %v", err, err)
		}
	}

	hasPass := strings.Contains(stdout, "All checks passed.")
	hasFail := strings.Contains(stdout, "failed.")
	if !hasPass && !hasFail {
		t.Errorf("stdout missing summary line; got:\n%s", stdout)
	}
}

// asErrCoder is a tiny errors.As-equivalent helper that avoids importing
// errors just for the one call site.
func asErrCoder(err error, dst *cli.ExitCoder) bool {
	if c, ok := err.(cli.ExitCoder); ok {
		*dst = c
		return true
	}
	return false
}
