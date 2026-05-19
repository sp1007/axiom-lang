package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func buildCLI(t *testing.T) string {
	exePath := filepath.Join(t.TempDir(), "axc.exe")
	cmd := exec.Command("go", "build", "-o", exePath, "./")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build axc: %v", err)
	}
	return exePath
}

func writeTempFile(t *testing.T, content string) string {
	path := filepath.Join(t.TempDir(), "test.ax")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckValidFile(t *testing.T) {
	axc := buildCLI(t)
	file := writeTempFile(t, "fn main():\n    let message = \"Hello\"\n")

	cmd := exec.Command(axc, "check", file)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		t.Errorf("expected exit 0, got %v. stderr: %s", err, stderr.String())
	}

	if !bytes.Contains(stderr.Bytes(), []byte("0 errors, 0 warnings")) {
		t.Errorf("expected success summary, got %s", stderr.String())
	}
}

func TestCheckTypeError(t *testing.T) {
	axc := buildCLI(t)
	file := writeTempFile(t, "fn main():\n    let x: i32 = \"hello\"\n")

	cmd := exec.Command(axc, "check", file)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		t.Errorf("expected exit 1, got 0")
	}

	if !bytes.Contains(stderr.Bytes(), []byte("type mismatch")) {
		t.Errorf("expected type mismatch error, got %s", stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("error: 1 errors, 0 warnings emitted")) {
		t.Errorf("expected 1 error summary, got %s", stderr.String())
	}
}

func TestCheckParseError(t *testing.T) {
	axc := buildCLI(t)
	file := writeTempFile(t, "fn main() -> \n")

	cmd := exec.Command(axc, "check", file)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		t.Errorf("expected exit 1, got 0")
	}

	if !bytes.Contains(stderr.Bytes(), []byte("expected type")) {
		t.Errorf("expected parser error, got %s", stderr.String())
	}
}
