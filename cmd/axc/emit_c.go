package main

import (
	"fmt"
	"os"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// runEmitC reads an AXIOM source file, parses it, and emits C11 code.
// Output goes to stdout by default, or to a file with --output.
func runEmitC(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axc emit-c <file.ax> [--output <file.c>]")
		return 1
	}

	filename := args[0]
	outputPath := "" // empty = stdout

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "axc: --output requires a filename")
				return 1
			}
		default:
			fmt.Fprintf(os.Stderr, "axc: unknown flag %q\n", args[i])
			return 1
		}
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

	// Combine and report diagnostics
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
			fmt.Fprintf(os.Stderr, "axc: %d error(s), aborting C generation\n", len(diags))
			return 1
		}
	}

	// Build type table and symbol table
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Generate C
	pipeline := cgen.NewPipeline(table, intern, symbols, tree)

	// Write output
	if outputPath == "" {
		// Write to stdout
		if err := pipeline.GenerateC(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "axc: codegen error: %v\n", err)
			return 1
		}
	} else {
		f, err := os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "axc: cannot create %s: %v\n", outputPath, err)
			return 1
		}
		defer f.Close()

		if err := pipeline.GenerateC(f); err != nil {
			fmt.Fprintf(os.Stderr, "axc: codegen error: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "axc: wrote C output to %s\n", outputPath)
	}

	return 0
}
