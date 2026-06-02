package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseManifest_Valid(t *testing.T) {
	manifestContent := `
[package]
name = "my_project"
version = "0.1.0"
authors = ["Compiler Team", "Systems Dev"]

[dependencies]
core = "0.1.0"
math = { git = "https://github.com/axiom-lang/math.git", tag = "v0.1.0" }
utils = { path = "../utils" }
`
	manifest, err := ParseManifest(manifestContent)
	if err != nil {
		t.Fatalf("unexpected error parsing manifest: %v", err)
	}

	if manifest.Package.Name != "my_project" {
		t.Errorf("expected package name 'my_project', got %q", manifest.Package.Name)
	}
	if manifest.Package.Version != "0.1.0" {
		t.Errorf("expected package version '0.1.0', got %q", manifest.Package.Version)
	}
	if len(manifest.Package.Authors) != 2 || manifest.Package.Authors[0] != "Compiler Team" || manifest.Package.Authors[1] != "Systems Dev" {
		t.Errorf("expected authors ['Compiler Team', 'Systems Dev'], got %q", manifest.Package.Authors)
	}

	if len(manifest.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(manifest.Dependencies))
	}

	// Verify dependencies
	var core, math, utils *Dependency
	for i := range manifest.Dependencies {
		d := &manifest.Dependencies[i]
		switch d.Name {
		case "core":
			core = d
		case "math":
			math = d
		case "utils":
			utils = d
		}
	}

	if core == nil || core.Version != "0.1.0" || core.Git != "" || core.Path != "" {
		t.Errorf("invalid core dependency: %+v", core)
	}
	if math == nil || math.Git != "https://github.com/axiom-lang/math.git" || math.Tag != "v0.1.0" || math.Path != "" {
		t.Errorf("invalid math dependency: %+v", math)
	}
	if utils == nil || utils.Path != "../utils" || utils.Git != "" {
		t.Errorf("invalid utils dependency: %+v", utils)
	}
}

func TestParseManifest_Invalid(t *testing.T) {
	invalidContent := `
[package]
name = my_project # Missing quotes
`
	_, err := ParseManifest(invalidContent)
	if err == nil {
		t.Error("expected error parsing invalid manifest but got none")
	}
}

func TestLockfileRoundtrip(t *testing.T) {
	lock := &Lockfile{
		Packages: []LockedPackage{
			{
				Name:    "core",
				Version: "0.1.0",
				Source:  "git+https://github.com/axiom-lang/core.git#v0.1.0",
				Commit:  "9a8b7c6d5e4f3d2c1b0a",
				Sha256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			{
				Name:    "math",
				Version: "0.2.0",
				Source:  "git+https://github.com/axiom-lang/math.git#4a1b2c3d",
				Commit:  "4a1b2c3d5e4f3d2c1b0a",
				Sha256:  "c0884da18c21867c2d1b26806f9c8f2b4852bb023d8df5b306b986895cd452ab",
			},
		},
	}

	serialized, err := SerializeLockfile(lock)
	if err != nil {
		t.Fatalf("unexpected error serializing lockfile: %v", err)
	}

	parsed, err := ParseLockfile(serialized)
	if err != nil {
		t.Fatalf("unexpected error parsing serialized lockfile: %v", err)
	}

	if len(parsed.Packages) != 2 {
		t.Fatalf("expected 2 packages in parsed lockfile, got %d", len(parsed.Packages))
	}

	for i := range lock.Packages {
		p1 := lock.Packages[i]
		p2 := parsed.Packages[i]
		if p1.Name != p2.Name || p1.Version != p2.Version || p1.Source != p2.Source || p1.Commit != p2.Commit || p1.Sha256 != p2.Sha256 {
			t.Errorf("roundtrip mismatch for package %d:\nexpected: %+v\ngot: %+v", i, p1, p2)
		}
	}
}

func TestParseLockfile_Invalid(t *testing.T) {
	invalidContent := `
[[package]]
name = "core"
invalid_syntax
`
	_, err := ParseLockfile(invalidContent)
	if err == nil {
		t.Error("expected error parsing invalid lockfile but got none")
	}
}

func TestGetCacheRootDir(t *testing.T) {
	root, err := GetCacheRootDir()
	if err != nil {
		t.Fatalf("unexpected error getting cache root: %v", err)
	}
	if !strings.HasSuffix(root, filepath.Join(".axiom", "cache")) {
		t.Errorf("expected cache root to end with .axiom/cache, got %q", root)
	}
}

func TestFetchDependency_Path(t *testing.T) {
	// Create a temp folder for mock dependency
	tmpDepDir, err := os.MkdirTemp("", "axiom-mock-dep-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDepDir)

	// Write mock files inside
	mockFile := filepath.Join(tmpDepDir, "math.ax")
	if err := os.WriteFile(mockFile, []byte("pub fn add(a: i32, b: i32) -> i32: return a + b"), 0644); err != nil {
		t.Fatalf("failed to write mock file: %v", err)
	}

	dep := Dependency{
		Name: "math",
		Path: tmpDepDir,
	}

	locked, err := FetchDependency(dep, "")
	if err != nil {
		t.Fatalf("unexpected error fetching local path dependency: %v", err)
	}

	if locked.Name != "math" {
		t.Errorf("expected package name 'math', got %q", locked.Name)
	}
	if locked.Version != "local" {
		t.Errorf("expected version 'local', got %q", locked.Version)
	}
	if locked.Source != "path+"+tmpDepDir {
		t.Errorf("expected source path, got %q", locked.Source)
	}
	if locked.Commit != "local" {
		t.Errorf("expected commit 'local', got %q", locked.Commit)
	}
	if len(locked.Sha256) != 64 {
		t.Errorf("expected 64-character SHA-256 hash, got %q", locked.Sha256)
	}
}

func TestResolveDependencies_Transitive(t *testing.T) {
	// Create a temp root dir
	tmpRootDir, err := os.MkdirTemp("", "axiom-mock-project-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpRootDir)

	// Create subfolder dep1
	dep1Dir := filepath.Join(tmpRootDir, "dep1")
	if err := os.MkdirAll(dep1Dir, 0755); err != nil {
		t.Fatalf("failed to create dep1 dir: %v", err)
	}
	// Write dep1 manifest declaring dep2
	dep1Manifest := `
[package]
name = "dep1"
version = "0.1.0"

[dependencies]
dep2 = { path = "../dep2" }
`
	if err := os.WriteFile(filepath.Join(dep1Dir, "axiom.toml"), []byte(dep1Manifest), 0644); err != nil {
		t.Fatalf("failed to write dep1 manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dep1Dir, "dep1.ax"), []byte("pub fn d1() {}"), 0644); err != nil {
		t.Fatalf("failed to write dep1 source: %v", err)
	}

	// Create subfolder dep2
	dep2Dir := filepath.Join(tmpRootDir, "dep2")
	if err := os.MkdirAll(dep2Dir, 0755); err != nil {
		t.Fatalf("failed to create dep2 dir: %v", err)
	}
	dep2Manifest := `
[package]
name = "dep2"
version = "0.2.0"
`
	if err := os.WriteFile(filepath.Join(dep2Dir, "axiom.toml"), []byte(dep2Manifest), 0644); err != nil {
		t.Fatalf("failed to write dep2 manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dep2Dir, "dep2.ax"), []byte("pub fn d2() {}"), 0644); err != nil {
		t.Fatalf("failed to write dep2 source: %v", err)
	}

	// Create root manifest
	rootManifestContent := fmt.Sprintf(`
[package]
name = "root"
version = "1.0.0"

[dependencies]
dep1 = { path = "%s" }
`, strings.ReplaceAll(dep1Dir, "\\", "/"))

	manifest, err := ParseManifest(rootManifestContent)
	if err != nil {
		t.Fatalf("failed to parse root manifest: %v", err)
	}

	lock, err := ResolveDependencies(manifest, tmpRootDir)
	if err != nil {
		t.Fatalf("failed to resolve dependencies: %v", err)
	}

	if len(lock.Packages) != 2 {
		t.Fatalf("expected 2 resolved packages, got %d", len(lock.Packages))
	}

	p1 := lock.Packages[0] // dep1 (alphabetical order)
	p2 := lock.Packages[1] // dep2

	if p1.Name != "dep1" || p2.Name != "dep2" {
		t.Errorf("expected sorted packages dep1, dep2; got %s, %s", p1.Name, p2.Name)
	}
}

func TestResolveDependencies_Circular(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp("", "axiom-mock-project-circular-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpRootDir)

	dep1Dir := filepath.Join(tmpRootDir, "dep1")
	_ = os.MkdirAll(dep1Dir, 0755)
	dep1Manifest := `
[package]
name = "dep1"
version = "0.1.0"

[dependencies]
dep2 = { path = "../dep2" }
`
	_ = os.WriteFile(filepath.Join(dep1Dir, "axiom.toml"), []byte(dep1Manifest), 0644)
	_ = os.WriteFile(filepath.Join(dep1Dir, "dep1.ax"), []byte(""), 0644)

	dep2Dir := filepath.Join(tmpRootDir, "dep2")
	_ = os.MkdirAll(dep2Dir, 0755)
	dep2Manifest := `
[package]
name = "dep2"
version = "0.2.0"

[dependencies]
dep1 = { path = "../dep1" }
`
	_ = os.WriteFile(filepath.Join(dep2Dir, "axiom.toml"), []byte(dep2Manifest), 0644)
	_ = os.WriteFile(filepath.Join(dep2Dir, "dep2.ax"), []byte(""), 0644)

	rootManifestContent := fmt.Sprintf(`
[package]
name = "root"
version = "1.0.0"

[dependencies]
dep1 = { path = "%s" }
`, strings.ReplaceAll(dep1Dir, "\\", "/"))

	manifest, err := ParseManifest(rootManifestContent)
	if err != nil {
		t.Fatalf("failed to parse root manifest: %v", err)
	}

	_, err = ResolveDependencies(manifest, tmpRootDir)
	if err == nil || !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("expected circular dependency error, got: %v", err)
	}
}

