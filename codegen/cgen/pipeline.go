package cgen

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// Pipeline orchestrates the full C code generation process for an AXIOM module.
// It ties together the DeclEmitter (declarations), StmtGen (statements), and
// ExprGen (expressions) to produce a complete, compilable C11 source file.
//
// Usage:
//
//	p := NewPipeline(table, intern, symbols, tree)
//	var buf bytes.Buffer
//	if err := p.GenerateC(&buf); err != nil { ... }
//	// buf now contains valid C11 source
//
//	// Optionally compile:
//	if err := p.CompileC("output", "generated.c"); err != nil { ... }
type Pipeline struct {
	table   *types.TypeTable
	intern  *ast.InternPool
	symbols *sema.SymbolTable
	tree    *ast.AstTree
}

// NewPipeline creates a Pipeline with the given compilation context.
// All parameters must be non-nil and represent a fully typed AST ready
// for code generation (i.e., after type checking and ownership analysis).
func NewPipeline(
	table *types.TypeTable,
	intern *ast.InternPool,
	symbols *sema.SymbolTable,
	tree *ast.AstTree,
) *Pipeline {
	return &Pipeline{
		table:   table,
		intern:  intern,
		symbols: symbols,
		tree:    tree,
	}
}

// GenerateC writes the complete C11 source for the module to w.
//
// The output structure is:
//  1. Declarations section (via DeclEmitter): #include, forwards, types, globals, prototypes
//  2. Function definitions with bodies (via StmtGen)
//
// Returns an error if writing to w fails.
func (p *Pipeline) GenerateC(w io.Writer) error {
	// Stage 1: Process declarations
	emitter := NewDeclEmitter(p.table, p.intern, p.symbols, p.tree)
	emitter.ProcessModule()

	// Stage 2: Emit declarations (#include, forward decls, types, globals, prototypes)
	emitter.EmitTo(w)

	// Stage 3: Emit function definitions with bodies
	p.emitFuncDefs(w)

	return nil
}

// emitFuncDefs walks all top-level function declarations in the AST and emits
// their C function definitions with bodies.
func (p *Pipeline) emitFuncDefs(w io.Writer) {
	root := p.tree.Node(0) // NodeProgram
	child := root.FirstChild
	for child != ast.NullIdx {
		node := p.tree.Node(child)
		if node.Kind == ast.NodeFuncDecl {
			p.emitFuncDef(w, child, node)
		}
		child = node.NextSibling
	}
}

// emitFuncDef emits a single function definition (signature + body).
func (p *Pipeline) emitFuncDef(w io.Writer, idx uint32, node *ast.AstNode) {
	// Extract function name
	nameText := string(p.tree.TokenText(node.TokenIdx))

	// Build return type and parameters from the symbol table
	retType := "void"
	var paramStrs []string

	symIdx := node.Payload
	if symIdx != 0 && int(symIdx) < len(p.symbols.Symbols) {
		sym := p.symbols.SymbolAt(symIdx)
		if sym.TypeID != 0 {
			entry := p.table.Entry(types.TypeID(sym.TypeID))
			if entry.Kind == types.KindFunction {
				queue := NewTypeDeclQueue()
				fi := p.table.FuncInfo(types.TypeID(sym.TypeID))
				retType = CTypeName(fi.Return, p.table, p.intern, queue)

				// Collect parameter names from AST children
				var paramNames []string
				c := node.FirstChild
				for c != ast.NullIdx {
					cn := p.tree.Node(c)
					if cn.Kind == ast.NodeParamDecl {
						pName := string(p.tree.TokenText(cn.TokenIdx))
						paramNames = append(paramNames, pName)
					}
					c = cn.NextSibling
				}

				// Build parameter strings
				for i, pt := range fi.Params {
					ctype := CTypeName(pt, p.table, p.intern, queue)
					pname := fmt.Sprintf("p%d", i)
					if i < len(paramNames) {
						pname = paramNames[i]
					}
					paramStrs = append(paramStrs, fmt.Sprintf("%s %s", ctype, pname))
				}
			}
		}
	}

	// Visibility prefix
	visibility := ""
	if node.Flags&ast.FlagIsPub == 0 && node.Flags&ast.FlagIsExtern == 0 {
		visibility = "static "
	}

	mangledName := MangleFuncName("", nameText)

	// Emit function signature
	paramsStr := "void"
	if len(paramStrs) > 0 {
		paramsStr = strings.Join(paramStrs, ", ")
	}

	fmt.Fprintf(w, "\n%s%s %s(%s) {\n", visibility, retType, mangledName, paramsStr)

	// Find the body block (last child that is a NodeBlock)
	bodyIdx := p.findFuncBody(node)
	if bodyIdx != ast.NullIdx {
		iw := NewIndentWriter(w)
		iw.Indent()
		queue := NewTypeDeclQueue()
		sg := NewStmtGen(iw, p.table, p.intern, p.symbols, p.tree, queue)
		sg.EmitFuncBody(bodyIdx)
	}

	fmt.Fprintln(w, "}")
}

// findFuncBody locates the body block node among a function's children.
// The body is typically the last NodeBlock child.
func (p *Pipeline) findFuncBody(node *ast.AstNode) uint32 {
	var bodyIdx uint32 = ast.NullIdx
	child := node.FirstChild
	for child != ast.NullIdx {
		cn := p.tree.Node(child)
		if cn.Kind == ast.NodeBlock {
			bodyIdx = child
		}
		child = cn.NextSibling
	}
	return bodyIdx
}

// CompileC invokes a C compiler to compile the generated C source file.
// outputPath is the desired binary output path.
// cSrcPath is the path to the generated .c source file.
//
// The function auto-detects an available C compiler (gcc, clang, or cl.exe on Windows).
// It compiles with -std=c11 and -Wall.
//
// Returns an error if no compiler is found or if compilation fails.
func (p *Pipeline) CompileC(outputPath string, cSrcPath string) error {
	compiler, err := DetectCCompiler()
	if err != nil {
		return err
	}
	return InvokeCompiler(compiler, outputPath, cSrcPath, nil)
}

// CompileOptions holds optional settings for the C compilation step.
type CompileOptions struct {
	OptLevel    string   // -O0, -O1, -O2, -O3 (default: -O0)
	Debug       bool     // emit debug info (-g -DAX_DEBUG)
	IncludeDirs []string // additional -I directories
	ExtraSrcs   []string // additional .c source files to compile (runtime, etc.)
}

// CompileCWithOptions invokes a C compiler with detailed options.
func (p *Pipeline) CompileCWithOptions(outputPath string, cSrcPath string, opts CompileOptions) error {
	compiler, err := DetectCCompiler()
	if err != nil {
		return err
	}

	args := buildCompilerArgs(compiler, outputPath, cSrcPath, opts)
	cmd := exec.Command(compiler, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler (%s) failed:\n%s", compiler, stderr.String())
	}
	return nil
}

// DetectCCompiler finds an available C compiler on the system.
// It checks for gcc, clang, and (on Windows) cl.exe in that order.
// Returns the compiler command name and nil on success, or an error if none found.
func DetectCCompiler() (string, error) {
	candidates := cCompilerCandidates()
	for _, cc := range candidates {
		if _, err := exec.LookPath(cc); err == nil {
			return cc, nil
		}
	}
	return "", fmt.Errorf(
		"no C compiler found in PATH; install gcc or clang\n"+
			"  searched: %s", strings.Join(candidates, ", "))
}

// cCompilerCandidates returns the list of C compiler names to try,
// in preference order. On Windows, cl.exe is also checked.
func cCompilerCandidates() []string {
	candidates := []string{"gcc", "clang"}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, "cl.exe")
	}
	return candidates
}

// InvokeCompiler runs the given C compiler with standard flags.
// includeDirs specifies additional -I include paths (may be nil).
func InvokeCompiler(compiler, outputPath, cSrcPath string, includeDirs []string) error {
	opts := CompileOptions{
		OptLevel:    "-O0",
		IncludeDirs: includeDirs,
	}
	args := buildCompilerArgs(compiler, outputPath, cSrcPath, opts)

	cmd := exec.Command(compiler, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler (%s) failed:\n%s", compiler, stderr.String())
	}
	return nil
}

// buildCompilerArgs constructs the command-line arguments for the C compiler.
func buildCompilerArgs(compiler, outputPath, cSrcPath string, opts CompileOptions) []string {
	// MSVC (cl.exe) uses different flag syntax
	if strings.Contains(compiler, "cl") {
		return buildMSVCArgs(outputPath, cSrcPath, opts)
	}
	return buildGCCClangArgs(outputPath, cSrcPath, opts)
}

// buildGCCClangArgs builds args for GCC/Clang compilers.
func buildGCCClangArgs(outputPath, cSrcPath string, opts CompileOptions) []string {
	args := []string{
		cSrcPath,
		"-o", outputPath,
		"-std=c11",
		"-Wall",
	}

	optLevel := opts.OptLevel
	if optLevel == "" {
		optLevel = "-O0"
	}
	args = append(args, optLevel)

	if opts.Debug {
		args = append(args, "-g", "-DAX_DEBUG")
	}

	for _, dir := range opts.IncludeDirs {
		args = append(args, "-I"+dir)
	}

	args = append(args, opts.ExtraSrcs...)

	return args
}

// buildMSVCArgs builds args for cl.exe (MSVC).
func buildMSVCArgs(outputPath, cSrcPath string, opts CompileOptions) []string {
	args := []string{
		cSrcPath,
		"/Fe:" + outputPath,
		"/std:c11",
		"/W3",
	}

	if opts.Debug {
		args = append(args, "/Zi", "/DAX_DEBUG")
	}

	for _, dir := range opts.IncludeDirs {
		args = append(args, "/I"+dir)
	}

	args = append(args, opts.ExtraSrcs...)

	return args
}

// GenerateCSourcePath generates a default .c file path from a module name.
// For example, "main" → "main.c", "math/vector" → "math_vector.c".
func GenerateCSourcePath(moduleName string) string {
	// Replace path separators with underscores
	safe := strings.ReplaceAll(moduleName, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	return safe + ".c"
}

// OutputBinaryName derives the default output binary name from the input path.
// It strips the .ax extension if present: "main.ax" → "main", "hello" → "hello".
// On Windows, it appends .exe.
func OutputBinaryName(inputPath string) string {
	name := inputPath
	if strings.HasSuffix(name, ".ax") {
		name = name[:len(name)-3]
	}
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(name, ".exe") {
			name += ".exe"
		}
	}
	return name
}
