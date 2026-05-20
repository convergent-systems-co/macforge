//go:build tools

// Package main: blank imports here lock dependency versions in go.mod /
// go.sum before they are imported in production code. The //go:build tools
// constraint excludes this file from the default build, so the binary
// doesn't link them in. Removed when every listed module is imported from
// real code.
package main

import (
	_ "github.com/urfave/cli/v3"
	_ "gopkg.in/yaml.v3"
)
