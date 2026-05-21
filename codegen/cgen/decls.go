package cgen

import (
	"fmt"
	"io"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// DeclEmitter accumulates C declarations from the typed AST and emits them
// in the correct order: forward declarations, struct definitions, sum type
// definitions, function prototypes, and global variable declarations.
//
// Usage:
//
//	emitter := NewDeclEmitter(table, intern, symbols, tree)
//	emitter.ProcessModule()
//	emitter.EmitTo(os.Stdout)
type DeclEmitter struct {
	table   *types.TypeTable
	intern  *ast.InternPool
	symbols *sema.SymbolTable
	tree    *ast.AstTree
	queue   *TypeDeclQueue

	forwards []string // struct forward declarations
	typedefs []string // struct/sum/slice definitions
	protos   []string // function prototypes
	globals  []string // global variable declarations
	hasMain  bool
}

// NewDeclEmitter creates a DeclEmitter with the given compilation context.
func NewDeclEmitter(
	table *types.TypeTable,
	intern *ast.InternPool,
	symbols *sema.SymbolTable,
	tree *ast.AstTree,
) *DeclEmitter {
	return &DeclEmitter{
		table:   table,
		intern:  intern,
		symbols: symbols,
		tree:    tree,
		queue:   NewTypeDeclQueue(),
	}
}

// ProcessModule walks all top-level declarations in the AST and collects
// the corresponding C declarations. Call this before EmitTo.
func (e *DeclEmitter) ProcessModule() {
	root := e.tree.Node(0) // NodeProgram
	child := root.FirstChild
	for child != ast.NullIdx {
		node := e.tree.Node(child)
		switch node.Kind {
		case ast.NodeStructDecl:
			e.processStruct(child, node)
		case ast.NodeFuncDecl:
			e.processFunc(child, node)
		case ast.NodeConstDecl:
			e.processConst(child, node)
		case ast.NodeTypeAliasDecl:
			e.processTypeAlias(child, node)
		}
		child = node.NextSibling
	}

	// Drain the type declaration queue for any types referenced
	// by function signatures or global declarations.
	e.drainTypeDecls()
}

// EmitTo writes the full declaration section to w.
// The output ordering is:
//  1. #include "ax_runtime.h"
//  2. Forward declarations for all structs
//  3. Type definitions (structs, sum types, slices)
//  4. Global variable declarations
//  5. Function prototypes
func (e *DeclEmitter) EmitTo(w io.Writer) {
	if e.hasMain {
		fmt.Fprintln(w, "#define AX_EMIT_MAIN")
	}
	fmt.Fprintln(w, `#include "ax_runtime.h"`)
	fmt.Fprintln(w, `#include "ax_stdlib.h"`)
	fmt.Fprintln(w)

	// Forward declarations
	if len(e.forwards) > 0 {
		fmt.Fprintln(w, "/* Forward declarations */")
		for _, fwd := range e.forwards {
			fmt.Fprintln(w, fwd)
		}
		fmt.Fprintln(w)
	}

	// Type definitions
	if len(e.typedefs) > 0 {
		fmt.Fprintln(w, "/* Type definitions */")
		for _, td := range e.typedefs {
			fmt.Fprintln(w, td)
		}
		fmt.Fprintln(w)
	}

	// Globals
	if len(e.globals) > 0 {
		fmt.Fprintln(w, "/* Global variables */")
		for _, g := range e.globals {
			fmt.Fprintln(w, g)
		}
		fmt.Fprintln(w)
	}

	// Function prototypes
	if len(e.protos) > 0 {
		fmt.Fprintln(w, "/* Function prototypes */")
		for _, p := range e.protos {
			fmt.Fprintln(w, p)
		}
		fmt.Fprintln(w)
	}
}

// processStruct processes a struct declaration node.
func (e *DeclEmitter) processStruct(idx uint32, node *ast.AstNode) {
	nameID := e.nodeNameID(idx)
	name := e.resolveName(nameID, idx)

	// Forward declaration
	e.forwards = append(e.forwards, fmt.Sprintf("struct ax_%s;", name))

	// Collect fields
	var fields []fieldInfo
	child := node.FirstChild
	for child != ast.NullIdx {
		childNode := e.tree.Node(child)
		if childNode.Kind == ast.NodeFieldDecl {
			fNameID := e.nodeNameID(child)
			fName := e.resolveName(fNameID, child)
			
			// Resolve Field's TypeID from the symbol table if available,
			// otherwise fallback to the payload directly (for mock tests).
			fTypeID := types.TypeID(0)
			symIdx := childNode.Payload
			if symIdx != 0 && e.symbols != nil && int(symIdx) < len(e.symbols.Symbols) && e.symbols.SymbolAt(symIdx).Kind == sema.SymField {
				fTypeID = types.TypeID(e.symbols.SymbolAt(symIdx).TypeID)
			} else {
				fTypeID = types.TypeID(childNode.Payload)
			}
			
			fCType := CTypeName(fTypeID, e.table, e.intern, e.queue)
			fields = append(fields, fieldInfo{name: fName, ctype: fCType})
		}
		child = childNode.NextSibling
	}

	// Full definition
	var b strings.Builder
	fmt.Fprintf(&b, "struct ax_%s {\n", name)
	for _, f := range fields {
		fmt.Fprintf(&b, "    %s %s;\n", f.ctype, f.name)
	}
	b.WriteString("};")
	e.typedefs = append(e.typedefs, b.String())
}

// processFunc processes a function declaration node and emits a prototype.
func (e *DeclEmitter) processFunc(idx uint32, node *ast.AstNode) {
	nameID := e.nodeNameID(idx)
	name := e.resolveName(nameID, idx)
	if name == "main" {
		e.hasMain = true
	}

	// Determine return type from the symbol's TypeID
	var retType string
	var paramStrs []string

	symIdx := node.Payload // Payload = symbol index for FuncDecl
	if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
		sym := e.symbols.SymbolAt(symIdx)
		if sym.TypeID != 0 {
			entry := e.table.Entry(types.TypeID(sym.TypeID))
			if entry.Kind == types.KindFunction {
				fi := e.table.FuncInfo(types.TypeID(sym.TypeID))
				retType = CTypeName(fi.Return, e.table, e.intern, e.queue)

				// Build params from AST children + function type
				paramStrs = e.buildFuncParams(node, fi)
			}
		}
	}

	if retType == "" {
		retType = "void"
	}

	// Visibility
	visibility := ""
	if name != "main" && node.Flags&ast.FlagIsPub == 0 && node.Flags&ast.FlagIsExtern == 0 {
		visibility = "static "
	}

	mangledName := MangleFuncName("", name) // module prefix empty for now

	if len(paramStrs) == 0 {
		e.protos = append(e.protos, fmt.Sprintf("%s%s %s(void);", visibility, retType, mangledName))
	} else {
		e.protos = append(e.protos, fmt.Sprintf("%s%s %s(%s);",
			visibility, retType, mangledName, strings.Join(paramStrs, ", ")))
	}
}

// processConst processes a const declaration.
func (e *DeclEmitter) processConst(idx uint32, node *ast.AstNode) {
	nameID := e.nodeNameID(idx)
	name := e.resolveName(nameID, idx)

	symIdx := node.Payload
	ctype := "ax_i32" // default
	if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
		sym := e.symbols.SymbolAt(symIdx)
		if sym.TypeID != 0 {
			ctype = CTypeName(types.TypeID(sym.TypeID), e.table, e.intern, e.queue)
		}
	}

	mangledName := MangleGlobalName("", name)
	e.globals = append(e.globals, fmt.Sprintf("extern const %s %s;", ctype, mangledName))
}

// processTypeAlias processes a type alias (sum type) declaration.
func (e *DeclEmitter) processTypeAlias(idx uint32, node *ast.AstNode) {
	// Type aliases that map to sum types need their tag enum + struct emitted
	symIdx := node.Payload
	if symIdx == 0 || int(symIdx) >= len(e.symbols.Symbols) {
		return
	}
	sym := e.symbols.SymbolAt(symIdx)
	if sym.TypeID == 0 {
		return
	}

	typeID := types.TypeID(sym.TypeID)
	entry := e.table.Entry(typeID)
	if entry.Kind == types.KindSum {
		decl := CTypeDecl(typeID, e.table, e.intern, e.queue)
		if decl != "" {
			e.typedefs = append(e.typedefs, decl)
		}
	}
}

// drainTypeDecls processes all types enqueued by CTypeName calls.
func (e *DeclEmitter) drainTypeDecls() {
	ids := e.queue.Drain()
	for _, id := range ids {
		entry := e.table.Entry(id)
		switch entry.Kind {
		case types.KindSlice:
			decl := CTypeDecl(id, e.table, e.intern, e.queue)
			if decl != "" {
				e.typedefs = append(e.typedefs, decl)
			}
		}
	}
}

// buildFuncParams extracts parameter names from AST and types from FuncType.
func (e *DeclEmitter) buildFuncParams(node *ast.AstNode, fi *types.FuncType) []string {
	// Walk AST children to get parameter names
	var paramNames []string
	child := node.FirstChild
	for child != ast.NullIdx {
		childNode := e.tree.Node(child)
		if childNode.Kind == ast.NodeParamDecl {
			pNameID := e.nodeNameID(child)
			paramNames = append(paramNames, e.resolveName(pNameID, child))
		}
		child = childNode.NextSibling
	}

	// Build param strings combining types and names
	params := make([]string, len(fi.Params))
	for i, pt := range fi.Params {
		ctype := CTypeName(pt, e.table, e.intern, e.queue)
		pname := fmt.Sprintf("p%d", i) // fallback name
		if i < len(paramNames) {
			pname = paramNames[i]
		}
		params[i] = fmt.Sprintf("%s %s", ctype, pname)
	}
	if fi.IsVariadic {
		params = append(params, "...")
	}
	return params
}

// nodeNameID extracts the NameID for a declaration node.
// For declarations, the name is the interned form of the node's primary token text.
func (e *DeclEmitter) nodeNameID(idx uint32) uint32 {
	node := e.tree.Node(idx)
	if node.Kind == ast.NodeFieldDecl {
		// For NodeFieldDecl, Payload might contain the TypeID, not symIdx.
		// Return 0 to fallback to the token text as the name.
		return 0
	}
	symIdx := node.Payload
	if symIdx != 0 && e.symbols != nil && int(symIdx) < len(e.symbols.Symbols) {
		sym := e.symbols.SymbolAt(symIdx)
		return sym.NameID
	}
	return node.Payload
}

// resolveName converts a NameID back to a string, falling back to TokenText if 0.
func (e *DeclEmitter) resolveName(nameID uint32, nodeIdx uint32) string {
	if nameID == 0 {
		node := e.tree.Node(nodeIdx)
		txt := string(e.tree.TokenText(node.TokenIdx))
		if txt != "" {
			return txt
		}
		return "_anon"
	}
	return e.intern.Get(nameID)
}

// fieldInfo is an intermediate struct for field data collection.
type fieldInfo struct {
	name  string
	ctype string
}

// MangleFuncName creates a C-safe mangled name for a function.
// moduleName can be empty for the current module.
func MangleFuncName(moduleName, funcName string) string {
	if moduleName == "" {
		return "ax_" + funcName
	}
	return "ax_" + moduleName + "_" + funcName
}

// MangleGlobalName creates a C-safe mangled name for a global variable.
func MangleGlobalName(moduleName, varName string) string {
	if moduleName == "" {
		return "ax_" + varName
	}
	return "ax_" + moduleName + "_" + varName
}
