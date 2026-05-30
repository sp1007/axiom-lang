// Package driver implements the AXIOM compiler driver.
// It orchestrates the full compilation pipeline from source to executable,
// coordinating the lexer, parser, semantic analysis, IR, optimization,
// and code generation stages.
//
// The driver provides a single entry point [Compile] that runs the entire
// frontend pipeline (lex → parse → resolve → infer → typecheck → effects →
// arena → ownership → escape → ctgc → alias) and returns a [CompileResult]
// containing all intermediate products needed for code generation.
//
// This decouples the pipeline logic from the CLI commands, making it
// reusable for tooling (LSP, formatter, package manager) and testing.
package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// CompileResult holds all intermediate products from the frontend pipeline.
// Consumers can inspect diagnostics and, if no errors occurred, proceed
// to code generation using the typed AST, symbol table, and type table.
type CompileResult struct {
	// Source is the original source bytes.
	Source []byte

	// Filename is the source file path (used for diagnostics).
	Filename string

	// Tokens produced by the lexer.
	Tokens []lexer.Token

	// LineTable maps byte offsets to (line, col) positions.
	LineTable *lexer.LineTable

	// Intern is the string intern pool shared across all stages.
	Intern *ast.InternPool

	// Tree is the parsed AST.
	Tree *ast.AstTree

	// Symbols is the symbol table populated by name resolution.
	Symbols *sema.SymbolTable

	// Types is the type table populated by type inference.
	Types *types.TypeTable

	// Infer is the inference engine instance (needed by some codegen paths).
	Infer *sema.InferenceEngine

	// Ownership is the ownership checker instance (holds FunctionGraphs/FunctionMoved
	// needed by escape analysis and CTGC).
	Ownership *sema.OwnershipChecker

	// Diags contains all diagnostics accumulated during compilation.
	Diags []diagnostics.Diagnostic

	// Phase records how far the pipeline progressed before stopping.
	Phase CompilePhase
}

// CompilePhase indicates how far the pipeline progressed.
type CompilePhase int

const (
	// PhaseLex indicates the pipeline completed lexing.
	PhaseLex CompilePhase = iota
	// PhaseParse indicates the pipeline completed parsing.
	PhaseParse
	// PhaseResolve indicates the pipeline completed name resolution.
	PhaseResolve
	// PhaseInfer indicates the pipeline completed type inference.
	PhaseInfer
	// PhaseTypeCheck indicates the pipeline completed type checking and effects.
	PhaseTypeCheck
	// PhaseOwnership indicates the pipeline completed ownership analysis.
	PhaseOwnership
	// PhaseFull indicates the pipeline completed all stages including
	// escape analysis, CTGC injection, and alias reuse.
	PhaseFull
)

// String returns a human-readable name for the phase.
func (p CompilePhase) String() string {
	switch p {
	case PhaseLex:
		return "lex"
	case PhaseParse:
		return "parse"
	case PhaseResolve:
		return "resolve"
	case PhaseInfer:
		return "infer"
	case PhaseTypeCheck:
		return "typecheck"
	case PhaseOwnership:
		return "ownership"
	case PhaseFull:
		return "full"
	default:
		return "unknown"
	}
}

// HasErrors returns true if any diagnostic has severity Error.
func (r *CompileResult) HasErrors() bool {
	for _, d := range r.Diags {
		if d.Severity == diagnostics.SeverityError {
			return true
		}
	}
	return false
}

// ResolvePositions fills in Line/Col for any diagnostics that only have Offset.
func (r *CompileResult) ResolvePositions() {
	if r.LineTable == nil {
		return
	}
	for i := range r.Diags {
		if r.Diags[i].Pos.Line == 0 {
			line, col := r.LineTable.LineCol(r.Diags[i].Pos.Offset)
			r.Diags[i].Pos.Line = line
			r.Diags[i].Pos.Col = col
		}
	}
}

// CompileOptions configures the compilation pipeline.
type CompileOptions struct {
	// StopAfter, if set, causes the pipeline to stop after the specified phase
	// even if no errors have occurred. This is useful for tools that only need
	// partial analysis (e.g., a syntax checker only needs PhaseParse).
	StopAfter CompilePhase

	// StopAfterSet indicates whether StopAfter was explicitly set.
	// If false, the pipeline runs to completion (PhaseFull).
	StopAfterSet bool
}

// Compile runs the full AXIOM frontend pipeline on the given source.
// It returns a CompileResult containing all intermediate products.
// The pipeline stops early if errors are encountered at any stage,
// or if StopAfter is set in the options.
//
// Even when errors occur, the CompileResult contains partial results
// up to the phase that succeeded.
func Compile(source []byte, filename string, opts *CompileOptions) *CompileResult {
	if opts == nil {
		opts = &CompileOptions{}
	}

	stopPhase := PhaseFull
	if opts.StopAfterSet {
		stopPhase = opts.StopAfter
	}

	result := &CompileResult{
		Source:   source,
		Filename: filename,
	}

	// --- Phase: Lex ---
	tokens, lt, lexDiags := lexer.Lex(source)
	result.Tokens = tokens
	result.LineTable = lt
	result.Diags = append(result.Diags, lexDiags...)
	result.Phase = PhaseLex

	if stopPhase == PhaseLex {
		return result
	}

	// --- Phase: Parse ---
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)
	result.Intern = intern
	result.Tree = tree
	result.Diags = append(result.Diags, parseDiags...)
	result.Phase = PhaseParse

	if result.HasErrors() || stopPhase == PhaseParse {
		return result
	}

	// --- Phase: Name Resolution ---
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)
	result.Types = table
	result.Symbols = symbols

	// Setup LazyResolver with native module loader
	cwd, _ := os.Getwd()
	loadedModules := make(map[string]bool)
	var lr *sema.LazyResolver
	loader := func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		modulePath := intern.Get(m.NameID)
		fmt.Printf("[DEBUG LOADER ENTER] modulePath=%s loaded=%v status=%d\n", modulePath, loadedModules[modulePath], m.Status)
		if loadedModules[modulePath] {
			if m.Status == sema.ModuleLoaded {
				return nil
			}
			return fmt.Errorf("module %s previously failed to compile", modulePath)
		}
		loadedModules[modulePath] = true

		// Find the file path for the module
		var path string
		if strings.HasPrefix(modulePath, "std.") || strings.Contains(modulePath, ".") {
			rel := strings.Replace(modulePath, ".", "/", -1) + ".ax"
			path = filepath.Join(cwd, rel)
		} else {
			path = filepath.Join(filepath.Dir(filename), modulePath+".ax")
		}

		// Read module file
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read imported module %s: %v", modulePath, err)
		}

		// Parse the module file
		tokens, _, lexDiags := lexer.Lex(source)
		modTree, parseDiags := parser.Parse(tokens, source, intern)
		m.AstTree = modTree
		fmt.Printf("[DEBUG LOADER AST] modulePath=%s rootFirstChild=%d\n", modulePath, modTree.Node(0).FirstChild)

		// Save the current scope stack and temporarily set it to only contain the global scope (0)
		savedStack := st.GetStack()
		st.SetStack([]uint32{0})
		defer st.SetStack(savedStack)

		// Create a separate resolver for the module to resolve its names
		modResolver := sema.NewNameResolver(modTree, intern, st, tt, lr)
		resolveErrs := modResolver.Resolve()
		
		var modDiags []diagnostics.Diagnostic
		modDiags = append(modDiags, lexDiags...)
		modDiags = append(modDiags, parseDiags...)
		modDiags = append(modDiags, resolveErrs...)

		// Run type inference on the module so that all functions and symbols have their types populated
		modInfer := sema.NewInferenceEngine(modTree, st, tt, nil)
		inferErrs := modInfer.Infer()
		modDiags = append(modDiags, inferErrs...)

		// Run type checking on the module so that all structs and sum types are registered in the TypeTable
		modTC := sema.NewTypeChecker(modTree, intern, st, tt, modInfer)
		tcErrs := modTC.Check()
		modDiags = append(modDiags, tcErrs...)

		// Temporary debugging: dump types
		os.WriteFile("types_dump.txt", []byte(tt.DumpTypes()), 0644)
		os.WriteFile("symbols_dump.txt", []byte(st.DumpSymbols()), 0644)

		hasErrors := false
		for _, d := range modDiags {
			if d.Severity == diagnostics.SeverityError {
				hasErrors = true
			}
		}
		if hasErrors {
			opts := diagnostics.DefaultFormatOptions()
			for _, d := range modDiags {
				fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, source, path, opts))
			}
			return fmt.Errorf("compilation errors in module %s", modulePath)
		}

		// Export all public symbols defined in this module
		root := modTree.Node(0)
		child := root.FirstChild
		for child != ast.NullIdx {
			childNode := modTree.Node(child)
			if childNode.Kind == ast.NodeFuncDecl || childNode.Kind == ast.NodeStructDecl || childNode.Kind == ast.NodeConstDecl || childNode.Kind == ast.NodeTypeAliasDecl || childNode.Kind == ast.NodeInterfaceDecl {
				if childNode.Flags&uint16(ast.FlagIsPub) != 0 || childNode.Flags&uint16(ast.FlagIsExtern) != 0 {
					symIdx := childNode.Payload
					if symIdx != 0 {
						sym := st.SymbolAt(symIdx)
						m.Exports[sym.NameID] = symIdx
						fmt.Printf("[MODULE EXPORT] Module=%s ExportSymbol=%s symIdx=%d\n", modulePath, intern.Get(sym.NameID), symIdx)
					}
				}
			}
			child = childNode.NextSibling
		}

		return nil
	}

	lr = sema.NewLazyResolver(symbols, table, loader)

	// Preload std.result as prelude
	preludeModules := []string{"std.result"}
	for _, pm := range preludeModules {
		pmID := intern.InternString(pm)
		lr.RegisterImport(pmID, "", 0, 0)
		if err := lr.PreloadModule(pmID); err != nil {
			fmt.Printf("Error preloading prelude module %s: %v\n", pm, err)
		}
	}

	resolver := sema.NewNameResolver(tree, intern, symbols, table, lr)
	resolveErrs := resolver.Resolve()
	result.Diags = append(result.Diags, resolveErrs...)
	result.Phase = PhaseResolve

	if result.HasErrors() || stopPhase == PhaseResolve {
		return result
	}

	// --- Phase: Type Inference ---
	mono := sema.NewMonomorphizer(tree, intern, symbols, table)
	infer := sema.NewInferenceEngine(tree, symbols, table, mono)
	inferErrs := infer.Infer()
	result.Infer = infer
	result.Diags = append(result.Diags, inferErrs...)
	result.Phase = PhaseInfer

	if result.HasErrors() || stopPhase == PhaseInfer {
		return result
	}

	// --- Phase: Type Checking & Effects ---
	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	typeErrs := tc.Check()
	result.Diags = append(result.Diags, typeErrs...)

	effects := sema.NewEffectChecker(tree, intern, symbols, table, infer)
	effectErrs := effects.Check()
	result.Diags = append(result.Diags, effectErrs...)
	result.Phase = PhaseTypeCheck

	if result.HasErrors() || stopPhase == PhaseTypeCheck {
		return result
	}

	// --- Phase: Arena + Ownership ---
	ap := sema.NewArenaPass(tree, intern, symbols)
	arenaDiags := ap.Process()
	result.Diags = append(result.Diags, arenaDiags...)

	oc := sema.NewOwnershipChecker(tree, intern, symbols, table)
	ownershipDiags := oc.Check()
	result.Diags = append(result.Diags, ownershipDiags...)
	result.Ownership = oc
	result.Phase = PhaseOwnership

	if result.HasErrors() || stopPhase == PhaseOwnership {
		return result
	}

	// --- Phase: Escape Analysis + CTGC + Alias Reuse (per function) ---
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
					ctgc.Types = table
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

	result.Phase = PhaseFull
	fmt.Printf("[DRIVER COMPILE END] symbolsAddr=%p lazyAddr=%p\n", result.Symbols, result.Symbols.LazyResolver)
	if result.Symbols != nil && result.Symbols.LazyResolver != nil {
		fmt.Printf("[DRIVER COMPILE END] modulesCount=%d\n", len(result.Symbols.LazyResolver.GetModules()))
	}
	return result
}
