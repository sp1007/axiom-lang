package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAxiomSync(t *testing.T) {
	// 1. Concatenate the AXIOM sync implementation and its tests
	workspaceDir := "../.." // relative to tests/codegen
	syncPath := filepath.Join(workspaceDir, "std/sync.ax")
	resultPath := filepath.Join(workspaceDir, "std/result.ax") // Mutex uses Option[T]
	testPath := filepath.Join(workspaceDir, "std/sync_test.ax")

	sourceBytes, err := concatenateAxiomFiles(resultPath, syncPath, testPath)
	if err != nil {
		t.Fatalf("failed to concatenate sync files: %v", err)
	}

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-sync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "sync_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	compErr := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath, "")
	// Copy generated C file for debugging
	if cBytes, err := os.ReadFile(binPath + ".c"); err == nil {
		_ = os.WriteFile("../../sync_test_debug.c", cBytes, 0644)
	}
	if compErr != nil {
		t.Fatalf("failed to compile AXIOM sync: %v", compErr)
	}


	// 3. Run the compiled sync test executable
	cmd := exec.Command(binPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running AXIOM sync test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	t.Logf("Sync Test Output:\n%s", output)

	if !strings.Contains(output, "All AXIOM-native synchronization tests passed!") {
		t.Errorf("expected test success message, got output:\n%s", output)
	}
}
