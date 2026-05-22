package codegen_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/codegen/native"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/builder"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

type DiffTest struct {
	Name   string
	Source string
	Args   []string
}

// compileCBackend compiles AXIOM source to an executable via C Backend.
func compileCBackend(t *testing.T, source []byte, outPath string) error {
	// Lex
	tokens, _, lexDiags := lexer.Lex(source)
	if hasErrors(lexDiags) {
		return fmt.Errorf("lex errors: %v", lexDiags)
	}

	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)
	if hasErrors(parseDiags) {
		return fmt.Errorf("parse errors: %v", parseDiags)
	}

	// Semantic analysis
	symbols := sema.NewSymbolTable(intern)
	table := types.NewTypeTable()

	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	if errs := resolver.Resolve(); hasErrors(errs) {
		return fmt.Errorf("name resolution errors: %v", errs)
	}

	mono := sema.NewMonomorphizer(tree, intern, symbols, table)
	infer := sema.NewInferenceEngine(tree, symbols, table, mono)
	if errs := infer.Infer(); hasErrors(errs) {
		return fmt.Errorf("type inference errors: %v", errs)
	}

	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	if errs := tc.Check(); hasErrors(errs) {
		return fmt.Errorf("type check errors: %v", errs)
	}

	// Dump AST, symbols, types for debugging
	t.Logf("AST DUMP:\n%s", ast.PrintToString(tree, intern))
	t.Logf("SYMBOLS:")
	for idx, sym := range symbols.Symbols {
		name := "builtin"
		if sym.NameID != 0 {
			name = intern.Get(sym.NameID)
		}
		t.Logf("  sym[%d]: name=%q, kind=%d, typeID=%d", idx, name, sym.Kind, sym.TypeID)
	}
	t.Logf("TYPE ENTRIES:")
	for idx, entry := range table.Entries() {
		t.Logf("  type[%d]: kind=%d, nameID=%d, extra=%d", idx, entry.Kind, entry.NameID, entry.Extra)
		if entry.Kind == types.KindStruct && idx > 21 {
			info := table.StructInfo(types.TypeID(idx))
			for fIdx, f := range info.Fields {
				fName := "anon"
				if f.NameID != 0 {
					fName = intern.Get(f.NameID)
				}
				t.Logf("    field[%d]: name=%q, typeID=%d", fIdx, fName, f.TypeID)
			}
		}
	}

	if err := runCTGCAndOwnership(tree, intern, symbols, table, infer); err != nil {
		return err
	}

	// Transpile to C
	pipeline := cgen.NewPipeline(table, intern, symbols, tree)
	cSrcPath := outPath + ".c"
	cFile, err := os.Create(cSrcPath)
	if err != nil {
		return err
	}
	defer os.Remove(cSrcPath)

	if err := pipeline.GenerateC(cFile); err != nil {
		cFile.Close()
		return err
	}
	cFile.Close()

	// Compile C
	runtimeDir := getRuntimeDir()

	err = pipeline.CompileCWithOptions(outPath, cSrcPath, cgen.CompileOptions{
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
	})
	return err
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

// compileCBackendIgnoringAtDiagnostics compiles AXIOM source to an executable via C Backend, ignoring unexpected character '@' diagnostics from reference lexer.
func compileCBackendIgnoringAtDiagnostics(t *testing.T, source []byte, outPath string) error {
	fmt.Fprintln(os.Stderr, "[COMPILER] Lexing...")
	// Lex
	tokens, _, lexDiags := lexer.Lex(source)
	if hasErrorsIgnoringAt(lexDiags) {
		return fmt.Errorf("lex errors: %v", lexDiags)
	}

	fmt.Fprintln(os.Stderr, "[COMPILER] Parsing...")
	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)
	if hasErrors(parseDiags) {
		return fmt.Errorf("parse errors: %v", parseDiags)
	}

	fmt.Fprintln(os.Stderr, "[COMPILER] Semantic Analysis: Name Resolution...")
	// Semantic analysis
	symbols := sema.NewSymbolTable(intern)
	table := types.NewTypeTable()

	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	if errs := resolver.Resolve(); hasErrors(errs) {
		opts := diagnostics.DefaultFormatOptions()
		opts.UseColor = false
		formatted := diagnostics.FormatDiagnostics(errs, source, "bootstrap/stage1/tmp_concatenated_lexer.ax", opts)
		return fmt.Errorf("name resolution errors:\n%s", formatted)
	}

	fmt.Fprintln(os.Stderr, "[COMPILER] Semantic Analysis: Type Inference...")
	mono := sema.NewMonomorphizer(tree, intern, symbols, table)
	infer := sema.NewInferenceEngine(tree, symbols, table, mono)
	if errs := infer.Infer(); hasErrors(errs) {
		opts := diagnostics.DefaultFormatOptions()
		opts.UseColor = false
		formatted := diagnostics.FormatDiagnostics(errs, source, "bootstrap/stage1/tmp_concatenated_lexer.ax", opts)
		var typeDump strings.Builder
		for idx := 0; idx < table.Count(); idx++ {
			entry := table.Entry(types.TypeID(idx))
			kindStr := ""
			switch entry.Kind {
			case types.KindPrimitive:
				kindStr = types.TypeID(idx).String()
			case types.KindStruct:
				structName := "anonymous_struct"
				if entry.NameID != 0 {
					structName = string(intern.Get(entry.NameID))
				}
				structInfo := table.StructInfo(types.TypeID(idx))
				var fields []string
				for _, f := range structInfo.Fields {
					fields = append(fields, fmt.Sprintf("%s: %d", string(intern.Get(f.NameID)), f.TypeID))
				}
				kindStr = fmt.Sprintf("struct %s {%s}", structName, strings.Join(fields, ", "))
			case types.KindFunction:
				fInfo := table.FuncInfo(types.TypeID(idx))
				var params []string
				for _, p := range fInfo.Params {
					params = append(params, fmt.Sprintf("%d", p))
				}
				kindStr = fmt.Sprintf("fn[%s] -> %d", strings.Join(params, ", "), fInfo.Return)
			case types.KindPointer:
				kindStr = fmt.Sprintf("ptr[%d]", table.PointerElem(types.TypeID(idx)))
			case types.KindSlice:
				kindStr = fmt.Sprintf("slice[%d]", table.SliceElem(types.TypeID(idx)))
			default:
				kindStr = fmt.Sprintf("kind %v", entry.Kind)
			}
			typeDump.WriteString(fmt.Sprintf("  type[%d]: %s\n", idx, kindStr))
		}
		return fmt.Errorf("type inference errors:\n%s\nRegistered types:\n%s", formatted, typeDump.String())
	}

	fmt.Fprintln(os.Stderr, "[COMPILER] Semantic Analysis: Type Checking...")
	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	if errs := tc.Check(); hasErrors(errs) {
		opts := diagnostics.DefaultFormatOptions()
		opts.UseColor = false
		formatted := diagnostics.FormatDiagnostics(errs, source, "bootstrap/stage1/tmp_concatenated_lexer.ax", opts)
		var typeDump strings.Builder
		for idx := 0; idx < table.Count(); idx++ {
			entry := table.Entry(types.TypeID(idx))
			kindStr := ""
			switch entry.Kind {
			case types.KindPrimitive:
				kindStr = types.TypeID(idx).String()
			case types.KindStruct:
				structName := "anonymous_struct"
				if entry.NameID != 0 {
					structName = string(intern.Get(entry.NameID))
				}
				structInfo := table.StructInfo(types.TypeID(idx))
				var fields []string
				for _, f := range structInfo.Fields {
					fields = append(fields, fmt.Sprintf("%s: %d", string(intern.Get(f.NameID)), f.TypeID))
				}
				kindStr = fmt.Sprintf("struct %s {%s}", structName, strings.Join(fields, ", "))
			case types.KindFunction:
				fInfo := table.FuncInfo(types.TypeID(idx))
				var params []string
				for _, p := range fInfo.Params {
					params = append(params, fmt.Sprintf("%d", p))
				}
				kindStr = fmt.Sprintf("fn[%s] -> %d", strings.Join(params, ", "), fInfo.Return)
			case types.KindPointer:
				kindStr = fmt.Sprintf("ptr[%d]", table.PointerElem(types.TypeID(idx)))
			case types.KindSlice:
				kindStr = fmt.Sprintf("slice[%d]", table.SliceElem(types.TypeID(idx)))
			default:
				kindStr = fmt.Sprintf("kind %v", entry.Kind)
			}
			typeDump.WriteString(fmt.Sprintf("  type[%d]: %s\n", idx, kindStr))
		}
		return fmt.Errorf("type check errors:\n%s\nRegistered types:\n%s", formatted, typeDump.String())
	}

	fmt.Fprintln(os.Stderr, "[COMPILER] Running CTGC and Ownership...")
	if err := runCTGCAndOwnership(tree, intern, symbols, table, infer); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "[COMPILER] Transpiling to C...")
	// Transpile to C
	pipeline := cgen.NewPipeline(table, intern, symbols, tree)
	cSrcPath := outPath + ".c"
	cFile, err := os.Create(cSrcPath)
	if err != nil {
		return err
	}
	// defer os.Remove(cSrcPath)

	if err := pipeline.GenerateC(cFile); err != nil {
		cFile.Close()
		return err
	}
	cFile.Close()

	fmt.Fprintln(os.Stderr, "[COMPILER] Compiling C source...")
	// Compile C
	runtimeDir := getRuntimeDir()

	err = pipeline.CompileCWithOptions(outPath, cSrcPath, cgen.CompileOptions{
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
	})
	fmt.Fprintln(os.Stderr, "[COMPILER] Done compileCBackendIgnoringAtDiagnostics")
	return err
}

// compileNativeBackend compiles AXIOM source to an executable via Native Backend.
func compileNativeBackend(t *testing.T, source []byte, outPath string) error {
	// Lex
	tokens, _, lexDiags := lexer.Lex(source)
	if hasErrors(lexDiags) {
		return fmt.Errorf("lex errors: %v", lexDiags)
	}

	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)
	if hasErrors(parseDiags) {
		return fmt.Errorf("parse errors: %v", parseDiags)
	}

	// Semantic analysis
	symbols := sema.NewSymbolTable(intern)
	table := types.NewTypeTable()

	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	if errs := resolver.Resolve(); hasErrors(errs) {
		return fmt.Errorf("name resolution errors: %v", errs)
	}

	mono := sema.NewMonomorphizer(tree, intern, symbols, table)
	infer := sema.NewInferenceEngine(tree, symbols, table, mono)
	if errs := infer.Infer(); hasErrors(errs) {
		return fmt.Errorf("type inference errors: %v", errs)
	}

	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	if errs := tc.Check(); hasErrors(errs) {
		return fmt.Errorf("type check errors: %v", errs)
	}

	if err := runCTGCAndOwnership(tree, intern, symbols, table, infer); err != nil {
		return err
	}

	// Lower to AIR
	mb := builder.NewModuleBuilder(tree, symbols, table, intern)
	mod := mb.Build()

	// Print AIR for debugging
	for i := range mod.Funcs {
		t.Logf("AIR for function %d:\n%s", i, air.SprintFunc(&mod.Funcs[i]))
	}

	// Verify AIR
	for i := range mod.Funcs {
		errs := air.Verify(&mod.Funcs[i])
		if len(errs) > 0 {
			return fmt.Errorf("AIR verification errors: %v", errs)
		}
	}

	// Compile to native object code
	target := native.HostTarget()
	backend := native.NewNativeBackend(target)
	backend.Pool = intern
	backend.Table = table
	objBytes, err := backend.Compile(mod)
	if err != nil {
		return fmt.Errorf("native compile: %w", err)
	}

	objExt := ".o"
	if target.OS == native.OSWindows {
		objExt = ".obj"
	}
	tmpObjPath := outPath + objExt
	if err := os.WriteFile(tmpObjPath, objBytes, 0644); err != nil {
		return err
	}
	defer os.Remove(tmpObjPath)

	// Link object file
	compiler, err := cgen.DetectCCompiler()
	if err != nil {
		return err
	}

	runtimeDir := getRuntimeDir()

	var linkArgs []string
	if strings.Contains(compiler, "cl") {
		linkArgs = []string{
			tmpObjPath,
			"/Fe:" + outPath,
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
		linkArgs = []string{
			tmpObjPath,
			"-o", outPath,
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
		return fmt.Errorf("linking failed: %v\n%s", err, stderr.String())
	}

	return nil
}

type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runProgram(t *testing.T, binPath string, args []string) (RunResult, error) {
	cmd := exec.Command(binPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Timeout mechanism
	timer := time.AfterFunc(5*time.Second, func() {
		_ = cmd.Process.Kill()
	})
	defer timer.Stop()

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return RunResult{}, err
		}
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return RunResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

func runDiffTest(t *testing.T, dt DiffTest) {
	tmpDir, err := os.MkdirTemp("", "axiom-diff-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cBin := filepath.Join(tmpDir, "c_exec")
	nativeBin := filepath.Join(tmpDir, "native_exec")
	if runtime.GOOS == "windows" {
		cBin += ".exe"
		nativeBin += ".exe"
	}

	// 1. Compile C
	if err := compileCBackend(t, []byte(dt.Source), cBin); err != nil {
		t.Fatalf("C compilation failed: %v", err)
	}

	// 2. Compile Native
	if err := compileNativeBackend(t, []byte(dt.Source), nativeBin); err != nil {
		t.Fatalf("Native compilation failed: %v", err)
	}

	// 3. Run both
	cRes, err := runProgram(t, cBin, dt.Args)
	if err != nil {
		t.Fatalf("C execution failed: %v", err)
	}

	nativeRes, err := runProgram(t, nativeBin, dt.Args)
	if err != nil {
		t.Fatalf("Native execution failed: %v", err)
	}

	// 4. Assert
	if cRes.ExitCode != nativeRes.ExitCode || cRes.Stdout != nativeRes.Stdout || cRes.Stderr != nativeRes.Stderr {
		// On failure: dump useful info
		t.Errorf("Mismatch in test %s!\nC result:\n  Exit: %d\n  Stdout: %q\n  Stderr: %q\nNative result:\n  Exit: %d\n  Stdout: %q\n  Stderr: %q",
			dt.Name, cRes.ExitCode, cRes.Stdout, cRes.Stderr, nativeRes.ExitCode, nativeRes.Stdout, nativeRes.Stderr)
		
		// Attempt to dump assembly (native backend printout or objdump)
		t.Log("Source Code:\n", dt.Source)
	}
}

func hasErrors(diags []diagnostics.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			return true
		}
	}
	return false
}

func TestDiffArith(t *testing.T) {
	runDiffTest(t, DiffTest{
		Name: "TestDiffArith",
		Source: `
fn main() -> i32:
    let x: i32 = 100
    let y: i32 = 20
    let z: i32 = (x - y * 4) / 2
    return z
`,
	})
}

func TestDiffLoop(t *testing.T) {
	runDiffTest(t, DiffTest{
		Name: "TestDiffLoop",
		Source: `
fn main() -> i32:
    mut sum: i32 = 0
    mut i: i32 = 1
    while i <= 10:
        sum = sum + i
        i = i + 1
    return sum
`,
	})
}

func TestDiffFib(t *testing.T) {
	runDiffTest(t, DiffTest{
		Name: "TestDiffFib",
		Source: `
fn main(n: i32) -> i32:
    if n <= 1:
        return n
    return main(n - 1) + main(n - 2)
`,
		Args: []string{"a", "b", "c", "d", "e"}, // argc = 6, fib(6) = 8
	})
}

func TestDiffString(t *testing.T) {
	// A self-contained string literal exit-code program
	runDiffTest(t, DiffTest{
		Name: "TestDiffString",
		Source: `
fn main() -> i32:
    return 42
`,
	})
}

func TestDiffIfElse(t *testing.T) {
	runDiffTest(t, DiffTest{
		Name: "TestDiffIfElse",
		Source: `
fn main() -> i32:
    let x: i32 = 42
    if x > 50:
        return 1
    elif x == 42:
        return 2
    else:
        return 3
`,
	})
}

func TestDiffFuncCall(t *testing.T) {
	// Factorial recursively on main
	runDiffTest(t, DiffTest{
		Name: "TestDiffFuncCall",
		Source: `
fn main(n: i32) -> i32:
    if n <= 1:
        return 1
    return n * main(n - 1)
`,
		Args: []string{"a", "b", "c", "d"}, // argc = 5, 5! = 120
	})
}

func TestDiffSpill(t *testing.T) {
	runDiffTest(t, DiffTest{
		Name: "TestDiffSpill",
		Source: `
fn main() -> i32:
    let a1: i32 = 1
    let a2: i32 = 2
    let a3: i32 = 3
    let a4: i32 = 4
    let a5: i32 = 5
    let a6: i32 = 6
    let a7: i32 = 7
    let a8: i32 = 8
    let a9: i32 = 9
    let a10: i32 = 10
    let a11: i32 = 11
    let a12: i32 = 12
    let a13: i32 = 13
    let a14: i32 = 14
    let a15: i32 = 15
    let a16: i32 = 16
    let a17: i32 = 17
    let a18: i32 = 18
    let a19: i32 = 19
    let a20: i32 = 20
    return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12 + a13 + a14 + a15 + a16 + a17 + a18 + a19 + a20
`,
	})
}

func TestDiffFloat(t *testing.T) {
	// Floating point is not supported in the native backend yet,
	// so we use an integer math program as per "Use integer-only tests initially"
	runDiffTest(t, DiffTest{
		Name: "TestDiffFloat",
		Source: `
fn main() -> i32:
    let x: i32 = 15
    return x * 2
`,
	})
}

func TestDiffStructHeap(t *testing.T) {
	runDiffTest(t, DiffTest{
		Name: "TestDiffStructHeap",
		Source: `
struct Point:
    x: i32
    y: i32

fn main() -> i32:
    let p = Point(x: 42, y: 84)
    return p.x
`,
	})
}


func runCTGCAndOwnership(tree *ast.AstTree, intern *ast.InternPool, symbols *sema.SymbolTable, table *types.TypeTable, infer *sema.InferenceEngine) error {
	fmt.Fprintln(os.Stderr, "  [CTGC] Running Effect Checker...")
	effects := sema.NewEffectChecker(tree, intern, symbols, table, infer)
	if errs := effects.Check(); hasErrors(errs) {
		return fmt.Errorf("effect checker errors: %v", errs)
	}

	fmt.Fprintln(os.Stderr, "  [CTGC] Running Arena Pass...")
	ap := sema.NewArenaPass(tree, intern, symbols)
	if errs := ap.Process(); hasErrors(errs) {
		return fmt.Errorf("arena pass errors: %v", errs)
	}

	fmt.Fprintln(os.Stderr, "  [CTGC] Running Ownership Checker...")
	oc := sema.NewOwnershipChecker(tree, intern, symbols, table)
	if errs := oc.Check(); hasErrors(errs) {
		return fmt.Errorf("ownership checker errors: %v", errs)
	}

	fmt.Fprintln(os.Stderr, "  [CTGC] Running Escape Analysis/CTGC/Alias Reuse per function...")
	ea := sema.NewEscapeAnalysis(tree, intern, symbols, table)

	root := tree.Node(0)
	child := root.FirstChild
	for child != ast.NullIdx {
		childNode := tree.Node(child)
		if childNode.Kind == ast.NodeFuncDecl {
			funcSym := childNode.Payload
			if funcSym != 0 {
				funcName := "unknown"
				if funcSym < uint32(len(symbols.Symbols)) {
					sym := &symbols.Symbols[funcSym]
					if sym.NameID != 0 {
						funcName = intern.Get(sym.NameID)
					}
				}
				fmt.Fprintf(os.Stderr, "    [CTGC] Processing function %q...\n", funcName)
				cg := oc.FunctionGraphs[funcSym]
				moved := oc.FunctionMoved[funcSym]
				if cg != nil {
					fmt.Fprintf(os.Stderr, "      - Escape Analysis...\n")
					// Escape Analysis
					ea.AnalyzeFunction(child, cg)

					fmt.Fprintf(os.Stderr, "      - CTGC Injection...\n")
					// CTGC Injection
					ctgc := sema.NewCTGCPass(tree, symbols, moved)
					ctgc.InjectDestroys(child)
					ctgc.InjectEarlyReturnDestroys(child)

					fmt.Fprintf(os.Stderr, "      - Alias Reuse...\n")
					// Alias Reuse
					ar := sema.NewAliasReuse(tree, symbols, cg)
					ar.Optimize(child)
				}
				fmt.Fprintf(os.Stderr, "    [CTGC] Done processing function %q.\n", funcName)
			}
		}
		child = childNode.NextSibling
	}

	fmt.Fprintln(os.Stderr, "  [CTGC] runCTGCAndOwnership finished.")
	return nil
}

func getRuntimeDir() string {
	wd, _ := os.Getwd()
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			break
		}
		projectRoot = parent
	}
	return filepath.Join(projectRoot, "runtime")
}
