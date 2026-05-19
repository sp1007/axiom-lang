//go:build debug

// Package assert provides debug-build invariant checking.
// Functions in this package are active ONLY when built with -tags debug.
// In release builds (the default), all functions are no-ops that compile
// to zero instructions.
//
// Usage: assert.Invariant(len(nodes) > 0, "nodes must not be empty")
//
// NOTE: This is NOT a general-purpose assertion library. It is intended
// for internal compiler invariant checking only. See docs/CONTRIBUTING.md
// for the no-panic policy in compiler passes.
package assert

import "fmt"

// Invariant panics if cond is false. Only active in debug builds.
// The panic message includes "INVARIANT VIOLATION:" prefix for easy identification.
func Invariant(cond bool, msg string, args ...any) {
	if !cond {
		panic(fmt.Sprintf("INVARIANT VIOLATION: "+msg, args...))
	}
}

// Unreachable panics with a message indicating unreachable code was reached.
// Only active in debug builds.
func Unreachable(msg string, args ...any) {
	panic(fmt.Sprintf("UNREACHABLE: "+msg, args...))
}
