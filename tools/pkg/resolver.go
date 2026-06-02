package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ResolveDependencies recursively walks the dependency graph starting from the root manifest,
// downloads and caches git-based dependencies, validates local path dependencies,
// and returns a fully populated Lockfile.
func ResolveDependencies(manifest *Manifest, baseDir string) (*Lockfile, error) {
	resolved := make(map[string]LockedPackage)
	active := make(map[string]bool)

	if err := resolveRecursive(manifest.Dependencies, baseDir, resolved, active); err != nil {
		return nil, err
	}

	// Sort alphabetically by package name for deterministic lockfile output
	var keys []string
	for k := range resolved {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lock := &Lockfile{}
	for _, k := range keys {
		lock.Packages = append(lock.Packages, resolved[k])
	}

	return lock, nil
}

func resolveRecursive(deps []Dependency, baseDir string, resolved map[string]LockedPackage, active map[string]bool) error {
	for _, dep := range deps {
		if active[dep.Name] {
			return fmt.Errorf("circular dependency detected on package %s", dep.Name)
		}

		// Fetch / locate the dependency
		locked, err := FetchDependency(dep, baseDir)
		if err != nil {
			return err
		}

		// Conflict check
		if existing, exists := resolved[dep.Name]; exists {
			// If we resolved a path dependency and now request it again, ensure the paths match
			if existing.Source != locked.Source {
				return fmt.Errorf("dependency conflict for package %q:\n  previously resolved to: %s\n  now requested as: %s",
					dep.Name, existing.Source, locked.Source)
			}
			continue
		}

		resolved[dep.Name] = locked

		// Resolve transitive dependencies if this package has an axiom.toml manifest
		depDir := dep.Path
		if dep.Git != "" {
			cacheRoot, err := GetCacheRootDir()
			if err != nil {
				return err
			}
			depDir = filepath.Join(cacheRoot, dep.Name, locked.Commit)
		} else if !filepath.IsAbs(depDir) {
			depDir = filepath.Join(baseDir, dep.Path)
		}

		manifestPath := filepath.Join(depDir, "axiom.toml")
		if _, err := os.Stat(manifestPath); err == nil {
			// Read and parse transitive manifest
			contentBytes, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to read manifest of transitive dependency %s: %v", dep.Name, err)
			}
			transitiveManifest, err := ParseManifest(string(contentBytes))
			if err != nil {
				return fmt.Errorf("failed to parse manifest of transitive dependency %s: %v", dep.Name, err)
			}

			active[dep.Name] = true
			if err := resolveRecursive(transitiveManifest.Dependencies, depDir, resolved, active); err != nil {
				return err
			}
			delete(active, dep.Name)
		}
	}
	return nil
}
