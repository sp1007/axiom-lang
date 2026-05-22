//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func hasErrors(diags []diagnostics.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			return true
		}
	}
	return false
}

func hasErrorsIgnoringAt(diags []diagnostics.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			if strings.Contains(d.Message, "unexpected character '@'") {
				continue
			}
			return true
		}
	}
	return false
}

func runCTGCAndOwnership(tree *ast.AstTree, intern *ast.InternPool, symbols *sema.SymbolTable, table *types.TypeTable, infer *sema.InferenceEngine) error {
	effects := sema.NewEffectChecker(tree, intern, symbols, table, infer)
	if errs := effects.Check(); hasErrors(errs) {
		return fmt.Errorf("effect checker errors: %v", errs)
	}

	ap := sema.NewArenaPass(tree, intern, symbols)
	if errs := ap.Process(); hasErrors(errs) {
		return fmt.Errorf("arena pass errors: %v", errs)
	}

	oc := sema.NewOwnershipChecker(tree, intern, symbols, table)
	if errs := oc.Check(); hasErrors(errs) {
		return fmt.Errorf("ownership checker errors: %v", errs)
	}

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
					ea.AnalyzeFunction(child, cg)
					ctgc := sema.NewCTGCPass(tree, symbols, moved)
					ctgc.InjectDestroys(child)
					ctgc.InjectEarlyReturnDestroys(child)
					ar := sema.NewAliasReuse(tree, symbols, cg)
					ar.Optimize(child)
				}
			}
		}
		child = childNode.NextSibling
	}

	return nil
}

func main() {
	axiomRoot := `d:\projects\compiler\Axiom`
	sourcePath := filepath.Join(axiomRoot, "bootstrap/stage1/tmp_concatenated_air.ax")
	outPath := filepath.Join(axiomRoot, "my_air_test.c")

	fmt.Println("Reading source file...")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Printf("failed to read source: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Lexing...")
	tokens, _, lexDiags := lexer.Lex(source)
	if hasErrorsIgnoringAt(lexDiags) {
		fmt.Printf("lex errors: %v\n", lexDiags)
		os.Exit(1)
	}

	fmt.Println("Parsing...")
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)
	if hasErrors(parseDiags) {
		fmt.Printf("parse errors: %v\n", parseDiags)
		os.Exit(1)
	}

	fmt.Println("Semantic Analysis: Name Resolution...")
	symbols := sema.NewSymbolTable(intern)
	table := types.NewTypeTable()

	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	if errs := resolver.Resolve(); hasErrors(errs) {
		fmt.Printf("name resolution errors: %v\n", errs)
		os.Exit(1)
	}

	fmt.Println("Semantic Analysis: Type Inference...")
	infer := sema.NewInferenceEngine(tree, symbols, table, nil)
	if errs := infer.Infer(); hasErrors(errs) {
		fmt.Printf("type inference errors: %v\n", errs)
		os.Exit(1)
	}

	fmt.Println("Semantic Analysis: Type Checking...")
	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	if errs := tc.Check(); hasErrors(errs) {
		fmt.Printf("type check errors: %v\n", errs)
		os.Exit(1)
	}

	fmt.Println("Running CTGC and Ownership...")
	if err := runCTGCAndOwnership(tree, intern, symbols, table, infer); err != nil {
		fmt.Printf("CTGC errors: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Transpiling to C...")
	pipeline := cgen.NewPipeline(table, intern, symbols, tree)
	cFile, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer cFile.Close()

	if err := pipeline.GenerateC(cFile); err != nil {
		fmt.Printf("code generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated clean C file at: %s\n", outPath)
}
