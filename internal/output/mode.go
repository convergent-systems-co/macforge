// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package output

// Mode selects between the human (colorized terminal) and JSON renderers.
type Mode int

const (
	// ModeHuman renders results in a colorized, terminal-friendly format.
	ModeHuman Mode = iota
	// ModeJSON renders results as one stable, schema-versioned JSON envelope per command.
	ModeJSON
)

// String reports the mode name as appears in --output flag values.
func (m Mode) String() string {
	switch m {
	case ModeJSON:
		return "json"
	default:
		return "human"
	}
}

// ParseMode is the inverse of String. Unknown inputs return ModeHuman.
func ParseMode(s string) Mode {
	if s == "json" {
		return ModeJSON
	}
	return ModeHuman
}
