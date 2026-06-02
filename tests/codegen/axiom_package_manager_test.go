package codegen_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/tools/pkg"
)

func TestEndToEndPackageManager(t *testing.T) {
	// 1. Create temporary directory for the workspace
	tmpWorkspace, err := os.MkdirTemp("", "axiom-pm-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpWorkspace)

	// Directories
	projectDir := filepath.Join(tmpWorkspace, "my_project")
	depDir := filepath.Join(tmpWorkspace, "mock_dep")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatalf("failed to create dep dir: %v", err)
	}

	// 2. Create dependency files
	depManifest := `
[package]
name = "mock_dep"
version = "0.2.5"
authors = ["Dependency Author"]
`
	depSource := `
pub fn add(a: i32, b: i32) -> i32:
    return a + b
`
	if err := os.WriteFile(filepath.Join(depDir, "axiom.toml"), []byte(depManifest), 0644); err != nil {
		t.Fatalf("failed to write dependency manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(depDir, "mock_dep.ax"), []byte(depSource), 0644); err != nil {
		t.Fatalf("failed to write dependency source: %v", err)
	}

	// 3. Create root project files
	rootManifest := fmt.Sprintf(`
[package]
name = "my_project"
version = "1.0.0"
authors = ["Project Owner"]

[dependencies]
mock_dep = { path = "%s" }
`, strings.ReplaceAll(depDir, "\\", "/"))

	rootSource := `
import mock_dep

pub fn main() -> i32:
    let x: i32 = mock_dep.add(20, 22)
    return x
`
	if err := os.WriteFile(filepath.Join(projectDir, "axiom.toml"), []byte(rootManifest), 0644); err != nil {
		t.Fatalf("failed to write root manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "main.ax"), []byte(rootSource), 0644); err != nil {
		t.Fatalf("failed to write root source: %v", err)
	}

	// 4. Run dependency resolution (mimicking 'axc get')
	t.Log("Running ResolveAndFetch on root manifest...")
	lock, err := pkg.ResolveAndFetch(filepath.Join(projectDir, "axiom.toml"))
	if err != nil {
		t.Fatalf("failed to resolve and fetch dependencies: %v", err)
	}

	if len(lock.Packages) != 1 {
		t.Fatalf("expected 1 resolved package, got %d", len(lock.Packages))
	}
	if lock.Packages[0].Name != "mock_dep" {
		t.Errorf("expected resolved package name 'mock_dep', got %q", lock.Packages[0].Name)
	}

	lockfilePath := filepath.Join(projectDir, "axiom.lock")
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		t.Fatal("expected axiom.lock to be generated but it was not found")
	}

	// 5. Compile the multi-package project using the full AXIOM compiler CLI
	// Find or build the compiler executable
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	// relative to tests/codegen is root
	rootPath := filepath.Join(cwd, "../..")
	
	axcBinPath := filepath.Join(tmpWorkspace, "axc")
	if filepath.Separator == '\\' {
		axcBinPath += ".exe"
	}

	t.Log("Building axc binary...")
	buildCmd := exec.Command("go", "build", "-o", axcBinPath, filepath.Join(rootPath, "cmd/axc"))
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build axc compiler: %v\nStderr: %s", err, buildStderr.String())
	}

	// Run 'axc build main.ax -o out.exe' from the projectDir
	t.Log("Compiling multi-package project via axc CLI...")
	outExePath := filepath.Join(projectDir, "my_project")
	if filepath.Separator == '\\' {
		outExePath += ".exe"
	}

	cmd := exec.Command(axcBinPath, "build", filepath.Join(projectDir, "main.ax"), "-o", outExePath)
	cmd.Dir = rootPath // Set working directory to the monorepo root so standard library matches relative cwd lookup
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to compile multi-package project: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// 6. Run the compiled executable and verify exit code 42
	t.Log("Running compiled multi-package executable...")
	runCmd := exec.Command(outExePath)
	var runStdout, runStderr bytes.Buffer
	runCmd.Stdout = &runStdout
	runCmd.Stderr = &runStderr

	err = runCmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("failed to run compiled executable: %v", err)
		}
	} else {
		exitCode = 0
	}

	t.Logf("Executable run completed. Exit code: %d, stdout: %q, stderr: %q", exitCode, runStdout.String(), runStderr.String())

	if exitCode != 42 {
		t.Errorf("expected program exit code 42, got %d", exitCode)
	}
}
