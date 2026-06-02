// Package pkg implements the AXIOM package manager.
// It handles dependency resolution, versioning, registry interaction,
// and local package caching for AXIOM projects.
package pkg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PackageInfo struct {
	Name    string
	Version string
	Authors []string
}

type Dependency struct {
	Name    string
	Version string // Optional semantic version range
	Git     string // Git URL if git-based
	Tag     string // Git tag
	Commit  string // Git commit hash
	Branch  string // Git branch
	Path    string // Local path if path-based
}

type Manifest struct {
	Package      PackageInfo
	Dependencies []Dependency
}

// ParseManifest parses an axiom.toml file content and returns a Manifest struct.
// It uses a custom, lightweight, zero-dependency TOML parser specifically tailored for axiom.toml.
func ParseManifest(content string) (*Manifest, error) {
	manifest := &Manifest{}
	lines := strings.Split(content, "\n")
	currentSection := ""

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip inline comments if they are not inside a string
		if idx := strings.Index(line, "#"); idx != -1 {
			// Basic check: only strip if outside of quotes
			inQuotes := false
			stripIdx := -1
			for i, ch := range line {
				if ch == '"' {
					inQuotes = !inQuotes
				}
				if ch == '#' && !inQuotes {
					stripIdx = i
					break
				}
			}
			if stripIdx != -1 {
				line = strings.TrimSpace(line[:stripIdx])
			}
		}

		if line == "" {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}

		// Key-value pairs
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			return nil, fmt.Errorf("line %d: invalid syntax (missing '=')", lineNum+1)
		}

		key := strings.TrimSpace(line[:eqIdx])
		value := strings.TrimSpace(line[eqIdx+1:])

		switch currentSection {
		case "package":
			if err := parsePackageField(manifest, key, value); err != nil {
				return nil, fmt.Errorf("line %d in [package]: %v", lineNum+1, err)
			}
		case "dependencies":
			dep, err := parseDependency(key, value)
			if err != nil {
				return nil, fmt.Errorf("line %d in [dependencies]: %v", lineNum+1, err)
			}
			manifest.Dependencies = append(manifest.Dependencies, dep)
		}
	}

	return manifest, nil
}

func parsePackageField(manifest *Manifest, key, value string) error {
	switch key {
	case "name":
		val, err := parseString(value)
		if err != nil {
			return err
		}
		manifest.Package.Name = val
	case "version":
		val, err := parseString(value)
		if err != nil {
			return err
		}
		manifest.Package.Version = val
	case "authors":
		val, err := parseStringArray(value)
		if err != nil {
			return err
		}
		manifest.Package.Authors = val
	}
	return nil
}

func parseDependency(name, valStr string) (Dependency, error) {
	dep := Dependency{Name: name}

	// Case 1: Simple version string, e.g. core = "0.1.0"
	if strings.HasPrefix(valStr, "\"") && strings.HasSuffix(valStr, "\"") {
		ver, err := parseString(valStr)
		if err != nil {
			return dep, err
		}
		dep.Version = ver
		return dep, nil
	}

	// Case 2: Inline table, e.g. core = { git = "...", tag = "..." }
	if strings.HasPrefix(valStr, "{") && strings.HasSuffix(valStr, "}") {
		tableContent := valStr[1 : len(valStr)-1]
		parts := splitCommaOutsideQuotes(tableContent)

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			eqIdx := strings.Index(part, "=")
			if eqIdx == -1 {
				return dep, fmt.Errorf("invalid inline table syntax: %q", part)
			}
			k := strings.TrimSpace(part[:eqIdx])
			vStr := strings.TrimSpace(part[eqIdx+1:])

			v, err := parseString(vStr)
			if err != nil {
				return dep, fmt.Errorf("invalid inline table value for %s: %v", k, err)
			}

			switch k {
			case "git":
				dep.Git = v
			case "tag":
				dep.Tag = v
			case "commit":
				dep.Commit = v
			case "branch":
				dep.Branch = v
			case "path":
				dep.Path = v
			case "version":
				dep.Version = v
			default:
				return dep, fmt.Errorf("unknown dependency field %q", k)
			}
		}

		if dep.Git == "" && dep.Path == "" {
			return dep, errors.New("dependency must declare either a 'git' or a 'path' source")
		}
		return dep, nil
	}

	return dep, fmt.Errorf("invalid dependency format: %q", valStr)
}

func parseString(val string) (string, error) {
	if len(val) < 2 || !strings.HasPrefix(val, "\"") || !strings.HasSuffix(val, "\"") {
		return "", fmt.Errorf("invalid string literal: %q", val)
	}
	return val[1 : len(val)-1], nil
}

func parseStringArray(val string) ([]string, error) {
	if len(val) < 2 || !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
		return nil, fmt.Errorf("invalid array literal: %q", val)
	}
	content := val[1 : len(val)-1]
	parts := splitCommaOutsideQuotes(content)
	var res []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		s, err := parseString(p)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

func splitCommaOutsideQuotes(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	for _, r := range s {
		if r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
		} else if r == ',' && !inQuotes {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())
	return parts
}

// GetCacheRootDir returns the path to the global .axiom cache root.
func GetCacheRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return filepath.Join(home, ".axiom", "cache"), nil
}

// FetchDependency pulls the specified dependency into the local cache (or validates local path)
// and returns a LockedPackage struct containing commit hashes and integrity checks.
func FetchDependency(dep Dependency, baseDir string) (LockedPackage, error) {
	locked := LockedPackage{Name: dep.Name}

	// Case 1: Local path dependency
	if dep.Path != "" {
		absPath := dep.Path
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(baseDir, dep.Path)
		}

		info, err := os.Stat(absPath)
		if err != nil || !info.IsDir() {
			return locked, fmt.Errorf("local path dependency %s not found or is not a directory: %v", absPath, err)
		}

		sha, err := CalculateDirSHA256(absPath)
		if err != nil {
			return locked, fmt.Errorf("failed to calculate integrity hash for path dependency %s: %v", absPath, err)
		}

		locked.Version = "local"
		locked.Source = fmt.Sprintf("path+%s", dep.Path)
		locked.Commit = "local"
		locked.Sha256 = sha
		return locked, nil
	}

	// Case 2: Git dependency
	if dep.Git != "" {
		cacheRoot, err := GetCacheRootDir()
		if err != nil {
			return locked, err
		}

		// We will temporarily clone the repo to resolve HEAD and commits
		tmpDir, err := os.MkdirTemp("", "axiom-pkg-clone-*")
		if err != nil {
			return locked, fmt.Errorf("failed to create temporary workspace: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Determine the ref to checkout
		ref := "main"
		if dep.Tag != "" {
			ref = dep.Tag
		} else if dep.Commit != "" {
			ref = dep.Commit
		} else if dep.Branch != "" {
			ref = dep.Branch
		}

		// Clone the repo
		if err := CloneRepo(dep.Git, tmpDir); err != nil {
			return locked, err
		}

		// Checkout requested ref
		if err := CheckoutRef(tmpDir, ref); err != nil {
			return locked, err
		}

		// Get the exact resolved commit hash
		commitHash, err := GetCommitHash(tmpDir)
		if err != nil {
			return locked, err
		}

		// Package cache path: ~/.axiom/cache/<package_name>/<commit_hash>/
		pkgCachePath := filepath.Join(cacheRoot, dep.Name, commitHash)
		locked.Version = dep.Version
		if locked.Version == "" {
			locked.Version = "0.0.0" // fallback
		}

		locked.Source = fmt.Sprintf("git+%s#%s", dep.Git, ref)
		locked.Commit = commitHash

		// Check if it is already in our global cache
		if _, err := os.Stat(pkgCachePath); err == nil {
			// Cache hit! Compute SHA-256 of the existing cache to verify integrity
			sha, err := CalculateDirSHA256(pkgCachePath)
			if err == nil {
				locked.Sha256 = sha
				return locked, nil
			}
			// If integrity check fails, clear and re-cache
			_ = os.RemoveAll(pkgCachePath)
		}

		// Cache miss: copy from temporary workspace to global cache
		if err := os.MkdirAll(filepath.Dir(pkgCachePath), 0755); err != nil {
			return locked, fmt.Errorf("failed to create cache directory: %v", err)
		}

		// Perform copy
		if err := copyDir(tmpDir, pkgCachePath); err != nil {
			return locked, fmt.Errorf("failed to copy dependency to cache: %v", err)
		}

		// Compute SHA-256 of the newly cached directory
		sha, err := CalculateDirSHA256(pkgCachePath)
		if err != nil {
			return locked, fmt.Errorf("failed to calculate integrity hash for cached package: %v", err)
		}
		locked.Sha256 = sha

		return locked, nil
	}

	return locked, errors.New("dependency has no git or path source specified")
}

// Helper to copy directories recursively
func copyDir(src, dst string) error {
	return filepathWalk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the git database to save space and avoid lock issues
		if info.IsDir() && info.Name() == ".git" {
			return filepathSkipDir
		}

		relPath, err := filepathRel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file contents
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = ioCopy(dstFile, srcFile)
		return err
	})
}

// Simple copy wrapper avoiding standard io package import issues
func ioCopy(dst io.Writer, src io.Reader) (int64, error) {
	// A simple chunked copy buffer
	buf := make([]byte, 32*1024)
	var written int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return written, er
		}
	}
	return written, nil
}

// ResolveAndFetch reads axiom.toml, resolves all direct and transitive dependencies,
// fetches them into global cache, and writes axiom.lock.
func ResolveAndFetch(manifestPath string) (*Lockfile, error) {
	contentBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest %s: %v", manifestPath, err)
	}

	manifest, err := ParseManifest(string(contentBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %v", err)
	}

	baseDir := filepath.Dir(manifestPath)
	lock, err := ResolveDependencies(manifest, baseDir)
	if err != nil {
		return nil, fmt.Errorf("dependency resolution failed: %v", err)
	}

	// Serialize and write lockfile
	serialized, err := SerializeLockfile(lock)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize lockfile: %v", err)
	}

	lockfilePath := filepath.Join(baseDir, "axiom.lock")
	if err := os.WriteFile(lockfilePath, []byte(serialized), 0644); err != nil {
		return nil, fmt.Errorf("failed to write lockfile %s: %v", lockfilePath, err)
	}

	return lock, nil
}

