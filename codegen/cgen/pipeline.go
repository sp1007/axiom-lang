package cgen

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"sort"
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

	// Emit entry point wrapper if main exists
	if emitter.hasMain {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "/* Entry point wrapper */")
		if emitter.mainHasArgs {
			fmt.Fprintln(w, "ax_i32 ax_main(ax_i32 argc, ax_u8** argv) {")
			if emitter.mainParamCount == 1 {
				fmt.Fprintln(w, "    return ax_main_usr(argc);")
			} else {
				fmt.Fprintln(w, "    return ax_main_usr(argc, argv);")
			}
			fmt.Fprintln(w, "}")
		} else {
			fmt.Fprintln(w, "ax_i32 ax_main(void) {")
			fmt.Fprintln(w, "    return ax_main_usr();")
			fmt.Fprintln(w, "}")
		}
	}

	// Emit bridge allocator functions for C runtime integration if ActorHeap is defined in the program
	hasActorHeap := false
	for idx := 0; idx < p.table.Count(); idx++ {
		entry := p.table.Entry(types.TypeID(idx))
		if entry.Kind == types.KindStruct && entry.NameID != 0 && p.intern.Get(entry.NameID) == "ActorHeap" {
			hasActorHeap = true
			break
		}
	}

	if hasActorHeap {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "/* Bridge allocator functions for C runtime integration */")
		fmt.Fprintln(w, "struct ax_ActorHeap;")
		fmt.Fprintln(w, "struct ax_ActorHeap* ax_ax_actor_heap_create(ax_u64 actor_id);")
		fmt.Fprintln(w, "void ax_ActorHeap_ax_actor_heap_destroy(struct ax_ActorHeap* heap);")
		fmt.Fprintln(w, "ax_u8* ax_ActorHeap_ax_actor_alloc(struct ax_ActorHeap* heap, ax_i64 user_size);")
		fmt.Fprintln(w, "void ax_ActorHeap_ax_actor_free(struct ax_ActorHeap* heap, ax_u8* user_ptr);")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "void* ax_actor_heap_create(unsigned long long actor_id) {")
		fmt.Fprintln(w, "    return (void*)ax_ax_actor_heap_create((ax_u64)actor_id);")
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w, "void ax_actor_heap_destroy(void* heap) {")
		fmt.Fprintln(w, "    ax_ActorHeap_ax_actor_heap_destroy((struct ax_ActorHeap*)heap);")
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w, "void* ax_actor_alloc(void* heap, size_t user_size) {")
		fmt.Fprintln(w, "    return (void*)ax_ActorHeap_ax_actor_alloc((struct ax_ActorHeap*)heap, (ax_i64)user_size);")
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w, "void ax_actor_free(void* heap, void* user_ptr) {")
		fmt.Fprintln(w, "    ax_ActorHeap_ax_actor_free((struct ax_ActorHeap*)heap, (ax_u8*)user_ptr);")
		fmt.Fprintln(w, "}")
	}

	return nil
}

// emitFuncDefs walks all top-level function declarations in the AST and emits
// their C function definitions with bodies.
func (p *Pipeline) emitFuncDefs(w io.Writer) {
	// 1. Process main tree
	p.emitTreeFuncDefs(w, p.tree)

	// 2. Process all loaded module trees
	if p.symbols != nil && p.symbols.LazyResolver != nil {
		var modKeys []uint32
		for k := range p.symbols.LazyResolver.GetModules() {
			modKeys = append(modKeys, k)
		}
		sort.Slice(modKeys, func(i, j int) bool {
			return modKeys[i] < modKeys[j]
		})
		for _, k := range modKeys {
			mod := p.symbols.LazyResolver.GetModules()[k]
			if mod.AstTree != nil {
				p.emitTreeFuncDefs(w, mod.AstTree)
			}
		}
	}

	// Process all instantiated generic functions/methods
	var instKeys []uint32
	for k := range p.symbols.InstantiatedToOriginalName {
		instKeys = append(instKeys, k)
	}
	// Sort to preserve perfect determinism
	sort.Slice(instKeys, func(i, j int) bool {
		return instKeys[i] < instKeys[j]
	})

	for _, instSymIdx := range instKeys {
		sym := p.symbols.SymbolAt(instSymIdx)
		if sym.DeclNode != 0 {
			declNode := p.tree.Node(sym.DeclNode)
			if declNode.Kind == ast.NodeFuncDecl {
				if declNode.Flags&ast.FlagIsExtern == 0 {
					if sym.TypeID != 0 && isGeneric(p.table, types.TypeID(sym.TypeID), make(map[types.TypeID]bool)) {
						continue
					}
					p.emitFuncDef(w, sym.DeclNode, declNode)
				}
			}
		}
	}
}

// emitFuncDef emits a single function definition (signature + body).
func (p *Pipeline) emitFuncDef(w io.Writer, idx uint32, node *ast.AstNode) {
	// Extract function name
	nameText := ""
	symIdx := node.Payload
	if symIdx != 0 && int(symIdx) < len(p.symbols.Symbols) {
		sym := p.symbols.SymbolAt(symIdx)
		if sym.Flags&sema.SymFlagGeneric != 0 || (sym.TypeID != 0 && isGeneric(p.table, types.TypeID(sym.TypeID), make(map[types.TypeID]bool))) {
			return
		}
		nameText = p.intern.Get(sym.NameID)
	} else if node.Payload != 0 {
		nameText = p.intern.Get(node.Payload)
	} else {
		nameText = string(p.tree.TokenText(node.TokenIdx))
	}

	// Build return type and parameters from the symbol table
	retType := "void"
	var paramStrs []string
	returnTypeID := types.TypeVoid

	if symIdx != 0 && int(symIdx) < len(p.symbols.Symbols) {
		sym := p.symbols.SymbolAt(symIdx)
		if sym.TypeID != 0 {
			entry := p.table.Entry(types.TypeID(sym.TypeID))
			if entry.Kind == types.KindFunction {
				queue := NewTypeDeclQueue()
				fi := p.table.FuncInfo(types.TypeID(sym.TypeID))
				returnTypeID = fi.Return
				retType = CTypeName(fi.Return, p.table, p.intern, queue)

				// Collect parameter names and flags from AST children
				var paramNames []string
				var paramFlags []uint16
				var paramIsFuncType []bool
				c := node.FirstChild
				for c != ast.NullIdx {
					cn := p.tree.Node(c)
					if cn.Kind == ast.NodeParamDecl {
						pSymIdx := cn.Payload
						pName := ""
						if pSymIdx != 0 && int(pSymIdx) < len(p.symbols.Symbols) {
							pSym := p.symbols.SymbolAt(pSymIdx)
							pName = p.intern.Get(pSym.NameID)
						} else if cn.Payload != 0 {
							pName = p.intern.Get(cn.Payload)
						} else {
							pName = string(p.tree.TokenText(cn.TokenIdx))
						}
						paramNames = append(paramNames, pName)
						paramFlags = append(paramFlags, cn.Flags)

						// Check if this parameter has a function type in AST
						isFuncType := false
						cc := cn.FirstChild
						for cc != ast.NullIdx {
							ccn := p.tree.Node(cc)
							if ccn.Kind == ast.NodeFuncType {
								isFuncType = true
								break
							}
							cc = ccn.NextSibling
						}
						paramIsFuncType = append(paramIsFuncType, isFuncType)
					}
					c = cn.NextSibling
				}

				// Build parameter strings using EmitParamDecl
				for i, pt := range fi.Params {
					pname := fmt.Sprintf("p%d", i)
					var flags uint16
					if i < len(paramNames) {
						pname = paramNames[i]
						flags = paramFlags[i]
						if paramIsFuncType[i] {
							pt = p.table.RegisterFunction(nil, types.TypeVoid, nil)
						}
					}
					paramStrs = append(paramStrs, EmitParamDecl(pname, pt, flags, p.table, p.intern, queue))
				}
			}
		}
	}

	// Visibility prefix
	visibility := ""
	if nameText != "main" && node.Flags&ast.FlagIsPub == 0 && node.Flags&ast.FlagIsExtern == 0 {
		visibility = "static "
	}

	mangledName := GetFuncMangledName(symIdx, nameText, p.table, p.symbols, p.intern)

	if nameText == "main" {
		retType = "ax_i32"
	}

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
		sg.ExprGen.ReturnType = returnTypeID
		sg.ExprGen.FuncNode = idx
		sg.EmitFuncBody(bodyIdx)
		if nameText == "main" {
			iw.Line("return 0;")
		}
	} else if nameText == "main" {
		fmt.Fprintln(w, "    return 0;")
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

	if runtime.GOOS == "windows" {
		args = append(args, "-Wl,--no-insert-timestamp")
	}

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

func (p *Pipeline) emitTreeFuncDefs(w io.Writer, tree *ast.AstTree) {
	oldTree := p.tree
	p.tree = tree
	defer func() { p.tree = oldTree }()

	root := tree.Node(0) // NodeProgram
	child := root.FirstChild
	for child != ast.NullIdx {
		node := tree.Node(child)
		if node.Kind == ast.NodeFuncDecl {
			symIdx := node.Payload
			name := ""
			if symIdx != 0 && int(symIdx) < len(p.symbols.Symbols) {
				sym := p.symbols.SymbolAt(symIdx)
				name = p.intern.Get(sym.NameID)
				fmt.Printf("[DEBUG-CGEN-EMIT] Encountered top fn %s: flagIsGeneric=%t, symFlagGeneric=%t, flags=%d\n", name, node.Flags&ast.FlagIsGeneric != 0, sym.Flags&sema.SymFlagGeneric != 0, node.Flags)
			} else {
				name = string(p.tree.TokenText(node.TokenIdx))
				fmt.Printf("[DEBUG-CGEN-EMIT] Encountered top fn %s (no sym): flagIsGeneric=%t, flags=%d\n", name, node.Flags&ast.FlagIsGeneric != 0, node.Flags)
			}
			if node.Flags&ast.FlagIsGeneric != 0 {
				child = node.NextSibling
				continue
			}
			if node.Flags&ast.FlagIsExtern == 0 {
				p.emitFuncDef(w, child, node)
			}
		} else if node.Kind == ast.NodeStructDecl {
			if node.Flags&ast.FlagIsGeneric != 0 {
				child = node.NextSibling
				continue
			}
			sChild := node.FirstChild
			for sChild != ast.NullIdx {
				sNode := tree.Node(sChild)
				if sNode.Kind == ast.NodeFuncDecl {
					if sNode.Flags&ast.FlagIsGeneric != 0 {
						sChild = sNode.NextSibling
						continue
					}
					if sNode.Flags&ast.FlagIsExtern == 0 {
						p.emitFuncDef(w, sChild, sNode)
					}
				}
				sChild = sNode.NextSibling
			}
		}
		child = node.NextSibling
	}
}
