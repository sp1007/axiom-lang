package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/codegen/native"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/builder"
)

// runBuild compiles an AXIOM source file to an executable.
// It supports compiling via either the C-Backend or the Native Backend.
func runBuild(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: axc build <file.ax> [-o <output>] [--target <triple>]")
		return 1
	}

	filename := ""
	outputPath := "" // auto-derive from filename
	targetTriple := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "axc: -o requires an output filename")
				return 1
			}
		case "--target":
			if i+1 < len(args) {
				targetTriple = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "axc: --target requires a target triple")
				return 1
			}
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "axc: unknown flag %q\n", args[i])
				return 1
			}
			if filename == "" {
				filename = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "axc: multiple input files are not supported: %q\n", args[i])
				return 1
			}
		}
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "axc: missing input filename")
		return 1
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
	tokens, lt, lexDiags := lexer.Lex(source)

	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)

	// Report diagnostics
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

	// Format and print all diagnostics
	if len(diags) > 0 {
		if lt != nil {
			for i := range diags {
				if diags[i].Pos.Line == 0 {
					line, col := lt.LineCol(diags[i].Pos.Offset)
					diags[i].Pos.Line = line
					diags[i].Pos.Col = col
				}
			}
		}
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

	// If --target is specified, use the native compilation pipeline
	if targetTriple != "" {
		target, err := native.ParseTarget(targetTriple)
		if err != nil {
			fmt.Fprintf(os.Stderr, "axc: invalid target triple: %v\n", err)
			return 1
		}

		// Build AIR
		mb := builder.NewModuleBuilder(tree, symbols, table, intern)
		mod := mb.Build()

		// Verify AIR
		for i := range mod.Funcs {
			errs := air.Verify(&mod.Funcs[i])
			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "axc: AIR verification error: %v\n", e)
				}
				return 1
			}
		}

		// Compile via native backend
		backend := native.NewNativeBackend(target)
		backend.Pool = intern
		backend.Table = table
		objBytes, err := backend.Compile(mod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "axc: native codegen error: %v\n", err)
			return 1
		}

		// Write object file to a temporary file
		tmpDir, err := os.MkdirTemp("", "axc-build-native-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "axc: cannot create temp dir: %v\n", err)
			return 1
		}
		defer os.RemoveAll(tmpDir)

		objExt := ".o"
		if target.OS == native.OSWindows {
			objExt = ".obj"
		}
		tmpObjPath := filepath.Join(tmpDir, "output"+objExt)
		if err := os.WriteFile(tmpObjPath, objBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "axc: cannot write temporary object file: %v\n", err)
			return 1
		}

		// Link object file with the runtime
		compiler, err := cgen.DetectCCompiler()
		if err != nil {
			fmt.Fprintf(os.Stderr, "axc: compiler not found for linking: %v\n", err)
			return 1
		}

		runtimeDir := findRuntimeDir()
		var linkArgs []string
		if strings.Contains(compiler, "cl") {
			// MSVC link arguments
			linkArgs = []string{
				tmpObjPath,
				"/Fe:" + outputPath,
				"/I" + runtimeDir,
				filepath.Join(runtimeDir, "axalloc", "axalloc.c"),
				filepath.Join(runtimeDir, "panic", "panic.c"),
				filepath.Join(runtimeDir, "ax_assert.c"),
				filepath.Join(runtimeDir, "ax_collections.c"),
				filepath.Join(runtimeDir, "ax_math.c"),
				filepath.Join(runtimeDir, "ax_print.c"),
				filepath.Join(runtimeDir, "ax_string_ops.c"),
			}
		} else {
			// GCC / Clang link arguments
			linkArgs = []string{
				tmpObjPath,
				"-o", outputPath,
				"-I" + runtimeDir,
				filepath.Join(runtimeDir, "axalloc", "axalloc.c"),
				filepath.Join(runtimeDir, "panic", "panic.c"),
				filepath.Join(runtimeDir, "ax_assert.c"),
				filepath.Join(runtimeDir, "ax_collections.c"),
				filepath.Join(runtimeDir, "ax_math.c"),
				filepath.Join(runtimeDir, "ax_print.c"),
				filepath.Join(runtimeDir, "ax_string_ops.c"),
			}
		}

		cmd := exec.Command(compiler, linkArgs...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "axc: linking failed: %v\n%s\n", err, stderr.String())
			return 1
		}

		fmt.Fprintf(os.Stderr, "axc: built native executable %s (target: %s)\n", outputPath, targetTriple)
		return 0
	}

	// Otherwise, fallback to C-Backend
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
			filepath.Join(runtimeDir, "ax_assert.c"),
			filepath.Join(runtimeDir, "ax_collections.c"),
			filepath.Join(runtimeDir, "ax_math.c"),
			filepath.Join(runtimeDir, "ax_print.c"),
			filepath.Join(runtimeDir, "ax_string_ops.c"),
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
