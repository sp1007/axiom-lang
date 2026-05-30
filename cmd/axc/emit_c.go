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
		case "-O0", "-O1", "-O2", "-O3", "-O", "--opt":
			// Accept optimization flags but ignore them for direct C transpilation
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
	var diags []diagnostics.Diagnostic
	diags = append(diags, lexDiags...)
	diags = append(diags, parseDiags...)

	// Build type table and symbol table
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Run semantic analysis if there are no parsing errors
	var hasParseErrors bool
	for _, d := range parseDiags {
		if d.Severity == diagnostics.SeverityError {
			hasParseErrors = true
			break
		}
	}

	if !hasParseErrors {
		// Name Resolution
		resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
		resolveErrs := resolver.Resolve()
		diags = append(diags, resolveErrs...)

		var hasSemaErrors bool
		for _, d := range diags {
			if d.Severity == diagnostics.SeverityError {
				hasSemaErrors = true
				break
			}
		}

		if !hasSemaErrors {
			// Type Inference & Checking
			infer := sema.NewInferenceEngine(tree, symbols, table, nil)
			inferErrs := infer.Infer()
			diags = append(diags, inferErrs...)

			var hasInferErrors bool
			for _, d := range diags {
				if d.Severity == diagnostics.SeverityError {
					hasInferErrors = true
					break
				}
			}

			if !hasInferErrors {
				tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
				typeErrs := tc.Check()
				diags = append(diags, typeErrs...)

				effects := sema.NewEffectChecker(tree, intern, symbols, table, infer)
				effectErrs := effects.Check()
				diags = append(diags, effectErrs...)

				// Run CTGC and Ownership pipeline if there are no errors so far
				var semaHasErrors bool
				for _, d := range diags {
					if d.Severity == diagnostics.SeverityError {
						semaHasErrors = true
						break
					}
				}
				if !semaHasErrors {
					// 1. Arena Pass
					ap := sema.NewArenaPass(tree, intern, symbols)
					arenaDiags := ap.Process()
					diags = append(diags, arenaDiags...)

					// 2. Ownership Checker
					oc := sema.NewOwnershipChecker(tree, intern, symbols, table)
					ownershipDiags := oc.Check()
					diags = append(diags, ownershipDiags...)

					var ownershipHasErrors bool
					for _, d := range diags {
						if d.Severity == diagnostics.SeverityError {
							ownershipHasErrors = true
							break
						}
					}

					if !ownershipHasErrors {
						// 3. Escape Analysis & CTGC & Alias Reuse per function
						ea := sema.NewEscapeAnalysis(tree, intern, symbols, table)

						root := tree.Node(0)
						child := root.FirstChild
						for child != ast.NullIdx {
							childNode := tree.Node(child)
							if childNode.Kind == ast.NodeFuncDecl {
								funcSym := childNode.Payload
								if funcSym != 0 {
									cg := oc.FunctionGraphs[funcSym]
									moved := oc.FunctionMoved[funcSym]
									if cg != nil {
										// Escape Analysis
										ea.AnalyzeFunction(child, cg)

										// CTGC Injection
										ctgc := sema.NewCTGCPass(tree, symbols, moved)
										ctgc.InjectDestroys(child)
										ctgc.InjectEarlyReturnDestroys(child)

										// Alias Reuse
										ar := sema.NewAliasReuse(tree, symbols, cg)
										ar.Optimize(child)
									}
								}
							}
							child = childNode.NextSibling
						}
					}
				}
			}
		}
	}

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
			fmt.Fprintf(os.Stderr, "axc: error(s), aborting C generation\n")
			return 1
		}
	}

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
