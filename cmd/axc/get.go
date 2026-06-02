package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/axiom-lang/axiom/tools/pkg"
)

// runGet implements the 'axc get' / 'axc install' subcommand.
// It resolves, downloads, caches, and locks all dependencies.
func runGet(args []string) int {
	manifestPath := "axiom.toml"
	
	// Parse optional manifest path
	for i := 0; i < len(args); i++ {
		if args[i] == "--manifest" && i+1 < len(args) {
			manifestPath = args[i+1]
			i++
		}
	}

	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc get: failed to get absolute path: %v\n", err)
		return 1
	}

	// Check if manifest exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "axc get: manifest file %s not found. Please create one first.\n", manifestPath)
		return 1
	}

	fmt.Printf("Resolving and fetching dependencies for AXIOM project: %s...\n", absPath)
	lock, err := pkg.ResolveAndFetch(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc get error: %v\n", err)
		return 1
	}

	fmt.Printf("\nSuccess! Resolved %d packages:\n", len(lock.Packages))
	for _, p := range lock.Packages {
		fmt.Printf("  - %s (%s) -> %s\n", p.Name, p.Version, p.Source)
	}
	fmt.Printf("Lockfile written to %s\n", filepath.Join(filepath.Dir(absPath), "axiom.lock"))

	return 0
}
