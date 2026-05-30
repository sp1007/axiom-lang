package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAxiomAllocator(t *testing.T) {
	// 1. Concatenate the AXIOM allocator implementation and its tests
	workspaceDir := "../.." // relative to tests/codegen
	allocPath := filepath.Join(workspaceDir, "std/mem/alloc.ax")
	testPath := filepath.Join(workspaceDir, "std/mem/alloc_test.ax")

	sourceBytes, err := concatenateAxiomFiles(allocPath, testPath)
	if err != nil {
		t.Fatalf("failed to concatenate allocator files: %v", err)
	}

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-alloc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "allocator_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath, ""); err != nil {
		t.Fatalf("failed to compile AXIOM allocator: %v", err)
	}

	// Copy generated C file for debugging
	if cBytes, err := os.ReadFile(binPath + ".c"); err == nil {
		_ = os.WriteFile("../../allocator_test_debug.c", cBytes, 0644)
	}

	// 3. Run the compiled allocator test executable
	cmd := exec.Command(binPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running AXIOM allocator test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	t.Logf("Allocator Test Output:\n%s", output)

	if !strings.Contains(output, "All AXIOM-native Allocator tests passed!") {
		t.Errorf("expected test success message, got output:\n%s", output)
	}
}
