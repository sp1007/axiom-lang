package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TestResult struct {
	FilePath string
	Category string
	Success  bool
	Output   string
}

func main() {
	axiomRoot := `d:\projects\compiler\Axiom`
	axcPath := filepath.Join(axiomRoot, "axc.exe")
	testsDir := filepath.Join(axiomRoot, "tests")

	var results []TestResult

	err := filepath.WalkDir(testsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".ax" {
			return nil
		}

		relPath, err := filepath.Rel(axiomRoot, path)
		if err != nil {
			relPath = path
		}

		// Group into categories based on parent directory
		parentDir := filepath.Base(filepath.Dir(path))
		category := parentDir
		if parentDir == "tests" {
			category = "root"
		}

		fmt.Printf("Compiling %s...\n", relPath)

		// Create a temporary output file for the executable
		tmpOut := filepath.Join(os.TempDir(), "axiom_test_temp.exe")
		os.Remove(tmpOut) // Clean up before running

		cmd := exec.Command(axcPath, "build", relPath, "-o", tmpOut)
		cmd.Dir = axiomRoot
		
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		runErr := cmd.Run()
		success := runErr == nil

		// Clean up the temp binary if created
		os.Remove(tmpOut)

		output := stderr.String()
		if output == "" {
			output = stdout.String()
		}

		results = append(results, TestResult{
			FilePath: relPath,
			Category: category,
			Success:  success,
			Output:   strings.TrimSpace(output),
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Generate Markdown report
	var md bytes.Buffer
	md.WriteString("# AXIOM Compiler Test Suite Evaluation Report\n\n")
	md.WriteString("This report summarizes the compilation results of all `*.ax` files found in the `tests/` directory and its subdirectories.\n\n")

	md.WriteString("## Summary\n\n")
	
	total := len(results)
	succeeded := 0
	failed := 0
	for _, r := range results {
		if r.Success {
			succeeded++
		} else {
			failed++
		}
	}

	md.WriteString(fmt.Sprintf("- **Total Files Scanned:** %d\n", total))
	md.WriteString(fmt.Sprintf("- **Successfully Compiled:** %d\n", succeeded))
	md.WriteString(fmt.Sprintf("- **Failed to Compile:** %d\n\n", failed))

	md.WriteString("### Breakdown by Category\n\n")
	categories := make(map[string][]TestResult)
	for _, r := range results {
		categories[r.Category] = append(categories[r.Category], r)
	}

	for cat, list := range categories {
		catSucceed := 0
		for _, r := range list {
			if r.Success {
				catSucceed++
			}
		}
		md.WriteString(fmt.Sprintf("- **%s**: %d/%d passed\n", cat, catSucceed, len(list)))
	}
	md.WriteString("\n---\n\n")

	md.WriteString("## Detailed Results\n\n")
	md.WriteString("| File Path | Category | Status | Expected Behavior / Details |\n")
	md.WriteString("|---|---|---|---|\n")

	for _, r := range results {
		status := "❌ Fail"
		if r.Success {
			status = "✅ Pass"
		}

		// Determine expected behavior
		expected := "Should Compile"
		base := filepath.Base(r.FilePath)
		if strings.HasPrefix(base, "err_") {
			expected = "Expected Semantic/Syntax Error"
		} else if r.Category == "root" {
			expected = "Future Language Spec (Not yet supported by Stage 1 Parser)"
		} else if strings.Contains(r.FilePath, "interface_missing") || strings.Contains(r.FilePath, "sum_type_nonexhaustive") || strings.Contains(r.FilePath, "async_await_outside") {
			expected = "Expected Semantic/Syntax Error"
		}

		detail := expected
		if !r.Success {
			// Extract first line of error or clean up
			errLines := strings.Split(r.Output, "\n")
			firstErr := ""
			for _, line := range errLines {
				if strings.Contains(line, "error[") || strings.Contains(line, "undefined") || strings.Contains(line, "panic") || strings.Contains(line, "failed") {
					firstErr = line
					break
				}
			}
			if firstErr == "" && len(errLines) > 0 {
				firstErr = errLines[0]
			}
			detail = fmt.Sprintf("%s (Error: `%s`)", expected, strings.ReplaceAll(firstErr, "|", "\\|"))
		}

		md.WriteString(fmt.Sprintf("| [%s](file:///%s) | %s | %s | %s |\n", base, filepath.Join(axiomRoot, r.FilePath), r.Category, status, detail))
	}

	reportPath := filepath.Join(axiomRoot, "scratch", "test_evaluation_report.md")
	err = os.WriteFile(reportPath, md.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error writing report: %v\n", err)
	} else {
		fmt.Printf("Report written to %s\n", reportPath)
	}
}
