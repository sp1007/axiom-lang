package pkg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneRepo clones a Git repository into the specified destination directory.
func CloneRepo(url, destDir string) error {
	// Ensure the parent directory of destDir exists
	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", parent, err)
	}

	// Run git clone
	cmd := exec.Command("git", "clone", url, destDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed for %s: %v\nStderr: %s", url, err, stderr.String())
	}

	return nil
}

// CheckoutRef checks out a specific reference (tag, commit, or branch) in the repository directory.
func CheckoutRef(repoDir, ref string) error {
	// First fetch all to ensure we have the reference
	fetchCmd := exec.Command("git", "fetch", "--all", "--tags")
	fetchCmd.Dir = repoDir
	_ = fetchCmd.Run() // non-fatal if offline but reference is cached

	// Run git checkout
	cmd := exec.Command("git", "checkout", ref)
	cmd.Dir = repoDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout %s failed in %s: %v\nStderr: %s", ref, repoDir, err, stderr.String())
	}

	return nil
}

// GetCommitHash returns the 40-character commit hash of the current HEAD in the repository.
func GetCommitHash(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse HEAD failed in %s: %v\nStderr: %s", repoDir, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
