package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestFmtCheckSubcommand(t *testing.T) {
	axc := buildCLI(t)

	// 1. Unformatted file check
	unformattedCode := `fn main() {
  let x=1+2
}`
	file := writeTempFile(t, unformattedCode)

	// Running with --check should fail (exit non-zero) since it needs formatting
	cmd := exec.Command(axc, "fmt", "--check", file)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		t.Fatalf("expected axc fmt --check to fail on unformatted file, but it succeeded. stdout: %s, stderr: %s", stdout.String(), stderr.String())
	}

	// 2. Perform the formatting write
	cmdWrite := exec.Command(axc, "fmt", file)
	if err := cmdWrite.Run(); err != nil {
		t.Fatalf("failed to format file in place: %v", err)
	}

	// Verify file was formatted in place
	formattedContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	expected := `fn main() {
    let x = 1 + 2
}
`
	if string(formattedContent) != expected {
		t.Errorf("file formatting in place failed.\nGot:\n%q\nWant:\n%q", string(formattedContent), expected)
	}

	// 3. Formatted file check should now succeed (exit 0)
	cmdCheckSuccess := exec.Command(axc, "fmt", "--check", file)
	if err := cmdCheckSuccess.Run(); err != nil {
		t.Errorf("expected axc fmt --check to succeed on formatted file, but failed: %v", err)
	}
}
