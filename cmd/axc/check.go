package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func runCheck(args []string) {
	flags := flag.NewFlagSet("check", flag.ExitOnError)
	warningsAsErrors := flags.Bool("warnings-as-errors", false, "Treat warnings as errors")
	noColor := flags.Bool("no-color", false, "Disable colored output")
	verbose := flags.Bool("verbose", false, "Print each pipeline stage")
	
	// -v alias
	flags.BoolVar(verbose, "v", false, "Print each pipeline stage")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if flags.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: axc check <file.ax> [flags]")
		os.Exit(1)
	}

	inputFile := flags.Arg(0)
	src, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc: error reading %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "[axc] checking %s\n", inputFile)
	}

	var allDiags []diagnostics.Diagnostic

	// Lexer
	if *verbose {
		fmt.Fprintln(os.Stderr, "[axc] stage: lexer")
	}
	tokens, lt, lexErrs := lexer.Lex(src)
	allDiags = append(allDiags, lexErrs...)

	// Parser
	if *verbose {
		fmt.Fprintln(os.Stderr, "[axc] stage: parser")
	}
	pool := ast.NewInternPool(1024)
	tree, parseErrs := parser.Parse(tokens, src, pool)
	allDiags = append(allDiags, parseErrs...)

	// Name Resolver
	symtab := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	
	if !hasFatal(parseErrs) {
		if *verbose {
			fmt.Fprintln(os.Stderr, "[axc] stage: name resolver")
		}
		resolver := sema.NewNameResolver(tree, pool, symtab, tt, nil)
		resolveErrs := resolver.Resolve()
		allDiags = append(allDiags, resolveErrs...)
	}

	// Type Inference & Checking
	if !hasFatal(allDiags) {
		if *verbose {
			fmt.Fprintln(os.Stderr, "[axc] stage: type inference")
		}
		infer := sema.NewInferenceEngine(tree, symtab, tt, nil)
		inferErrs := infer.Infer()
		allDiags = append(allDiags, inferErrs...)

		if !hasFatal(inferErrs) {
			if *verbose {
				fmt.Fprintln(os.Stderr, "[axc] stage: type checking")
			}
			tc := sema.NewTypeChecker(tree, pool, symtab, tt, infer)
			typeErrs := tc.Check()
			allDiags = append(allDiags, typeErrs...)
			
			if *verbose {
				fmt.Fprintln(os.Stderr, "[axc] stage: effect checking")
			}
			effects := sema.NewEffectChecker(tree, pool, symtab, tt, infer)
			effectErrs := effects.Check()
			allDiags = append(allDiags, effectErrs...)

			// Run CTGC and Ownership pipeline diagnostics
			var checkHasErrors bool
			for _, d := range allDiags {
				if d.Severity == diagnostics.SeverityError {
					checkHasErrors = true
					break
				}
			}
			if !checkHasErrors {
				if *verbose {
					fmt.Fprintln(os.Stderr, "[axc] stage: arena pass")
				}
				ap := sema.NewArenaPass(tree, pool, symtab)
				arenaDiags := ap.Process()
				allDiags = append(allDiags, arenaDiags...)

				if *verbose {
					fmt.Fprintln(os.Stderr, "[axc] stage: ownership checker")
				}
				oc := sema.NewOwnershipChecker(tree, pool, symtab, tt)
				ownershipDiags := oc.Check()
				allDiags = append(allDiags, ownershipDiags...)

				var ownershipHasErrors bool
				for _, d := range allDiags {
					if d.Severity == diagnostics.SeverityError {
						ownershipHasErrors = true
						break
					}
				}

				if !ownershipHasErrors {
					if *verbose {
						fmt.Fprintln(os.Stderr, "[axc] stage: escape analysis & ctgc & alias reuse")
					}
					ea := sema.NewEscapeAnalysis(tree, pool, symtab, tt)

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
									ctgc := sema.NewCTGCPass(tree, symtab, moved)
									ctgc.InjectDestroys(child)
									ctgc.InjectEarlyReturnDestroys(child)

									// Alias Reuse
									ar := sema.NewAliasReuse(tree, symtab, cg)
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

	allDiags = deduplicate(allDiags)
	
	if lt != nil {
		for i := range allDiags {
			if allDiags[i].Pos.Line == 0 {
				line, col := lt.LineCol(allDiags[i].Pos.Offset)
				allDiags[i].Pos.Line = line
				allDiags[i].Pos.Col = col
			}
		}
	}

	sortDiags(allDiags)

	fmtOpts := diagnostics.DefaultFormatOptions()
	fmtOpts.UseColor = !*noColor

	output := diagnostics.FormatDiagnostics(allDiags, src, inputFile, fmtOpts)
	if len(allDiags) > 0 {
		fmt.Fprint(os.Stderr, output)
	}

	errors, warnings := countDiags(allDiags)
	if errors > 0 || warnings > 0 {
		fmt.Fprintf(os.Stderr, "error: %d errors, %d warnings emitted\n", errors, warnings)
	} else {
		fmt.Fprintf(os.Stderr, "axc: 0 errors, 0 warnings in %s\n", inputFile)
	}

	if errors > 0 || (*warningsAsErrors && warnings > 0) {
		os.Exit(1)
	}
	os.Exit(0)
}

func hasFatal(diags []diagnostics.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			return true
		}
	}
	return false
}

func countDiags(diags []diagnostics.Diagnostic) (int, int) {
	errs, warns := 0, 0
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityWarning {
			warns++
		} else {
			errs++
		}
	}
	return errs, warns
}

func deduplicate(diags []diagnostics.Diagnostic) []diagnostics.Diagnostic {
	seen := make(map[string]bool)
	var res []diagnostics.Diagnostic
	for _, d := range diags {
		key := fmt.Sprintf("%d:%d:%d", d.Pos.Line, d.Pos.Col, d.Code)
		if !seen[key] {
			seen[key] = true
			res = append(res, d)
		}
	}
	return res
}

func sortDiags(diags []diagnostics.Diagnostic) {
	sort.Slice(diags, func(i, j int) bool {
		if diags[i].Pos.Line != diags[j].Pos.Line {
			return diags[i].Pos.Line < diags[j].Pos.Line
		}
		if diags[i].Pos.Col != diags[j].Pos.Col {
			return diags[i].Pos.Col < diags[j].Pos.Col
		}
		return diags[i].Severity > diags[j].Severity // Fatal > Error > Warning
	})
}
