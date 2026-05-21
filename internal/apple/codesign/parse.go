// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package codesign

import "strings"

// parseDisplay extracts identifier, team ID, authority, and timestamp from
// `codesign --display --verbose=4` stderr.
//
// Apple writes display info to stderr (not stdout). Example:
//
//	Executable=/path/to/MyApp.app/Contents/MacOS/MyApp
//	Identifier=com.example.MyApp
//	TeamIdentifier=XYZ1234567
//	Authority=Developer ID Application: ACME Inc. (XYZ1234567)
//	Timestamp=Apr 21, 2026 at 14:30:23
func parseDisplay(stderr string) DisplayInfo {
	var info DisplayInfo
	for _, line := range strings.Split(stderr, "\n") {
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		switch k {
		case "Identifier":
			info.Identifier = v
		case "TeamIdentifier":
			info.TeamID = v
		case "Authority":
			if info.Authority == "" {
				info.Authority = v
			}
		case "Timestamp":
			info.Timestamp = v
		}
	}
	return info
}

func splitKV(line string) (string, string, bool) {
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}
