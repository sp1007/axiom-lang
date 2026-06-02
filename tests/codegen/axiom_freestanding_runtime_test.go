package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAxiomFreestandingRuntime(t *testing.T) {
	workspaceDir := "../.." // relative to tests/codegen
	syscallPath := filepath.Join(workspaceDir, "bootstrap/runtime/syscall.ax")
	panicPath := filepath.Join(workspaceDir, "bootstrap/runtime/panic.ax")
	genrefPath := filepath.Join(workspaceDir, "bootstrap/runtime/genref.ax")
	axallocPath := filepath.Join(workspaceDir, "bootstrap/runtime/axalloc.ax")
	schedulerStubPath := filepath.Join(workspaceDir, "bootstrap/runtime/scheduler_stub.ax")
	testPath := filepath.Join(workspaceDir, "bootstrap/runtime/runtime_test.ax")

	sourceBytes, err := concatenateAxiomFiles(syscallPath, panicPath, genrefPath, axallocPath, schedulerStubPath, testPath)
	if err != nil {
		t.Fatalf("failed to concatenate freestanding runtime files: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "axiom-freestanding-runtime-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "freestanding_runtime_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendFreestanding(t, sourceBytes, binPath, ""); err != nil {
		t.Fatalf("failed to compile AXIOM freestanding runtime: %v", err)
	}

	// Copy generated C file for debugging
	if cBytes, err := os.ReadFile(binPath + ".c"); err == nil {
		_ = os.WriteFile("../../freestanding_runtime_test_debug.c", cBytes, 0644)
	}

	cmd := exec.Command(binPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("running AXIOM freestanding runtime test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	t.Logf("Freestanding Runtime Test Output:\n%s", output)

	if !strings.Contains(output, "All AXIOM-native Freestanding Runtime tests passed!") {
		t.Errorf("expected test success message, got output:\n%s", output)
	}
}
