package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAxiomCollections(t *testing.T) {
	// 1. Concatenate the AXIOM result and collections implementation and its tests
	workspaceDir := "../.." // relative to tests/codegen
	resultPath := filepath.Join(workspaceDir, "std/result.ax") // collections uses Option[T]
	collectionsPath := filepath.Join(workspaceDir, "std/collections.ax")
	testPath := filepath.Join(workspaceDir, "std/collections_test.ax")

	sourceBytes, err := concatenateAxiomFiles(resultPath, collectionsPath, testPath)
	if err != nil {
		t.Fatalf("failed to concatenate collections files: %v", err)
	}

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-collections-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "collections_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath, ""); err != nil {
		t.Fatalf("failed to compile AXIOM collections: %v", err)
	}

	// Copy generated C file for debugging
	if cBytes, err := os.ReadFile(binPath + ".c"); err == nil {
		_ = os.WriteFile("../../collections_test_debug.c", cBytes, 0644)
	}

	// 3. Run the compiled collections test executable
	cmd := exec.Command(binPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running AXIOM collections test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	t.Logf("Collections Test Output:\n%s", output)

	if !strings.Contains(output, "All AXIOM-native collections tests passed!") {
		t.Errorf("expected test success message, got output:\n%s", output)
	}
}
