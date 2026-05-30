package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAxiomProcess(t *testing.T) {
	workspaceDir := "../.." // relative to tests/codegen
	resultPath := filepath.Join(workspaceDir, "std/result.ax")
	collectionsPath := filepath.Join(workspaceDir, "std/collections.ax")
	osPath := filepath.Join(workspaceDir, "std/os.ax")
	processPath := filepath.Join(workspaceDir, "std/process.ax")
	testPath := filepath.Join(workspaceDir, "std/process_test.ax")

	sourceBytes, err := concatenateAxiomFiles(resultPath, collectionsPath, osPath, processPath, testPath)
	if err != nil {
		t.Fatalf("failed to concatenate process files: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "axiom-process-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "process_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	compErr := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath, "")
	// Copy generated C file for debugging
	if cBytes, err := os.ReadFile(binPath + ".c"); err == nil {
		_ = os.WriteFile("../../process_test_debug.c", cBytes, 0644)
	}
	if compErr != nil {
		t.Fatalf("failed to compile AXIOM process: %v", compErr)
	}

	// Run the compiled process test executable
	cmd := exec.Command(binPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running AXIOM process test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	t.Logf("Process Test Output:\n%s", output)

	if !strings.Contains(output, "All AXIOM-native process tests passed!") {
		t.Errorf("expected test success message, got output:\n%s", output)
	}
}
