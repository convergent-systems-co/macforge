package shell

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/polliard/macheim/internal/config"
)

// AppendIfMissing appends line to rcPath only when the exact line is not
// already present anywhere in the file. The file is created with mode
// 0644 if missing. The append uses O_APPEND so concurrent writers from
// other tools cannot truncate each other's output.
//
// Idempotency is checked by exact-line comparison after splitting on
// '\n'. Whitespace and case are NOT normalized — "eval \"$(brew shellenv)\""
// and "eval $(brew shellenv)" are treated as distinct lines because they
// behave differently when sourced. Callers wanting fuzzy match should
// canonicalize line before calling.
//
// rt.DryRun prints "[dry-run] append <line> to <rcPath>" to os.Stderr and
// returns nil without touching the filesystem. A nil rt is treated as
// the zero value (DryRun false).
func AppendIfMissing(rt *config.Runtime, rcPath, line string) error {
	if rt != nil && rt.DryRun {
		_, _ = fmt.Fprintf(os.Stderr, "[dry-run] append %q to %s\n", line, rcPath)
		return nil
	}

	content, err := os.ReadFile(rcPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", rcPath, err)
	}

	for _, existing := range strings.Split(string(content), "\n") {
		if existing == line {
			return nil
		}
	}

	f, err := os.OpenFile(rcPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", rcPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("write %s: %w", rcPath, err)
	}
	return nil
}
