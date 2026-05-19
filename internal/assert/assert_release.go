//go:build !debug

// Package assert provides debug-build invariant checking.
// In release builds (the default), all functions are no-ops.
package assert

// Invariant is a no-op in release builds.
func Invariant(_ bool, _ string, _ ...any) {}

// Unreachable is a no-op in release builds.
func Unreachable(_ string, _ ...any) {}
