package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// runBuild compiles an AXIOM source file to an executable via the C-Backend.
//
// Pipeline: source -> lex -> parse -> sema -> emit-c -> gcc -> executable
func runBuild(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axc build <file.ax> [-o <output>]")
		return 1
	}

	filename := args[0]
	outputPath := "" // auto-derive from filename

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "axc: -o requires an output filename")
				return 1
			}
		default:
			fmt.Fprintf(os.Stderr, "axc: unknown flag %q\n", args[i])
			return 1
		}
	}

	// Auto-derive output path: "main.ax" -> "main" (+ ".exe" on Windows)
	if outputPath == "" {
		outputPath = cgen.OutputBinaryName(filepath.Base(filename))
	}

	// Read source
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc: cannot read %s: %v\n", filename, err)
		return 1
	}

	// Lex
	tokens, _, lexDiags := lexer.Lex(source)

	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)

	// Report diagnostics
	diags := append(lexDiags, parseDiags...)
	if len(diags) > 0 {
		opts := diagnostics.DefaultFormatOptions()
		hasErrors := false
		for _, d := range diags {
			fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, source, filename, opts))
			if d.Severity == diagnostics.SeverityError {
				hasErrors = true
			}
		}
		if hasErrors {
			return 1
		}
	}

	// Build type table and symbol table
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Generate C to a temp file
	pipeline := cgen.NewPipeline(table, intern, symbols, tree)

	// Create temp C source file
	cSrcPath := cgen.GenerateCSourcePath(filepath.Base(filename))
	tmpDir, err := os.MkdirTemp("", "axc-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc: cannot create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	cFilePath := filepath.Join(tmpDir, cSrcPath)
	cFile, err := os.Create(cFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc: cannot create %s: %v\n", cFilePath, err)
		return 1
	}

	if err := pipeline.GenerateC(cFile); err != nil {
		cFile.Close()
		fmt.Fprintf(os.Stderr, "axc: codegen error: %v\n", err)
		return 1
	}
	cFile.Close()

	// Compile C to executable
	runtimeDir := findRuntimeDir()
	if err := pipeline.CompileCWithOptions(outputPath, cFilePath, cgen.CompileOptions{
		IncludeDirs: []string{runtimeDir},
		ExtraSrcs: []string{
			filepath.Join(runtimeDir, "axalloc", "axalloc.c"),
			filepath.Join(runtimeDir, "panic", "panic.c"),
		},
		Debug: true,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "axc: C compilation failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "axc: built %s\n", outputPath)
	return 0
}

// findRuntimeDir locates the AXIOM runtime directory.
// Searches relative to the executable, then falls back to common locations.
func findRuntimeDir() string {
	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		candidates := []string{
			filepath.Join(dir, "runtime"),
			filepath.Join(dir, "..", "runtime"),
			filepath.Join(dir, "..", "..", "runtime"),
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				return c
			}
		}
	}

	// Try environment variable
	if rtDir := os.Getenv("AXIOM_RUNTIME"); rtDir != "" {
		return rtDir
	}

	// Fall back to working directory
	return "runtime"
}

// isWindows returns true if running on Windows.
func isWindows() bool {
	return os.PathSeparator == '\\'
}
