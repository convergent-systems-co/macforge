package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/polliard/macheim/internal/config"
)

// Prompt reads a y/n response from r and returns true only for y / Y /
// yes / YES (case-insensitive). It returns false on EOF, empty input,
// and anything else — "no" is always the safe default for mutating ops.
//
// rt.Yes short-circuits to (true, nil) without reading from r, so
// non-interactive runs (--yes) skip the prompt entirely. The prompt
// message is printed to os.Stderr so the response can be piped through
// stdout in scripts without contaminating it.
//
// A read error other than EOF is returned to the caller; EOF is treated
// as "no" because an unexpected EOF on stdin (closed pipe, broken tty)
// must not be interpreted as consent to a destructive operation.
func Prompt(rt *config.Runtime, r io.Reader, msg string) (bool, error) {
	if rt != nil && rt.Yes {
		return true, nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "%s [y/N] ", msg)

	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
