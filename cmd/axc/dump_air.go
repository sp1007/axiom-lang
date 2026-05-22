package main

import (
	"fmt"
	"os"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/builder"
)

// runDumpAIR compiles an AXIOM source file and prints its AIR representation.
//
// Pipeline: source -> lex -> parse -> (sema) -> build AIR -> verify -> print
func runDumpAIR(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axc dump-air <file.ax> [--verify] [--no-verify]")
		return 1
	}

	filename := args[0]
	verify := true

	for _, arg := range args[1:] {
		switch arg {
		case "--verify":
			verify = true
		case "--no-verify":
			verify = false
		default:
			fmt.Fprintf(os.Stderr, "axc: unknown flag %q\n", arg)
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

	// Report diagnostics
	diags := append(lexDiags, parseDiags...)
	hasErrors := false
	if len(diags) > 0 {
		opts := diagnostics.DefaultFormatOptions()
		for _, d := range diags {
			fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, source, filename, opts))
			if d.Severity == diagnostics.SeverityError {
				hasErrors = true
			}
		}
	}
	if hasErrors {
		return 1
	}

	// Build type table and symbol table
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Semantic analysis
	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	if errs := resolver.Resolve(); hasErrorsDiags(errs) {
		opts := diagnostics.DefaultFormatOptions()
		for _, d := range errs {
			fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, source, filename, opts))
		}
		return 1
	}

	infer := sema.NewInferenceEngine(tree, symbols, table, nil)
	if errs := infer.Infer(); hasErrorsDiags(errs) {
		opts := diagnostics.DefaultFormatOptions()
		for _, d := range errs {
			fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, source, filename, opts))
		}
		return 1
	}

	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	if errs := tc.Check(); hasErrorsDiags(errs) {
		opts := diagnostics.DefaultFormatOptions()
		for _, d := range errs {
			fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, source, filename, opts))
		}
		return 1
	}

	// Build AIR
	mb := builder.NewModuleBuilder(tree, symbols, table, intern)
	mod := mb.Build()

	// Verify (optional)
	if verify {
		for i := range mod.Funcs {
			errs := air.Verify(&mod.Funcs[i])
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "verify: %s\n", e.Error())
			}
		}
	}

	// Print
	air.PrintModule(os.Stdout, mod)

	return 0
}

func hasErrorsDiags(diags []diagnostics.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			return true
		}
	}
	return false
}

