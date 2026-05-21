//go:build tools

// Package main: blank imports here lock dependency versions in go.mod /
// go.sum before they are imported in production code. The //go:build tools
// constraint excludes this file from the default build, so the binary
// doesn't link them in. Removed when every listed module is imported from
// real code.
//
// Currently empty: every previously listed module is now imported from real
// code (urfave/cli/v3 from cmd/, yaml.v3 from internal/config/). The file is
// retained as a placeholder for the next dependency that needs version
// pinning before its first real consumer.
package main
