package cgen

import (
	"fmt"
	"io"
	"sort"
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
	emitted map[types.TypeID]bool

	forwards []string // struct forward declarations
	typedefs []string // struct/sum/slice definitions
	protos   []string // function prototypes
	globals  []string // global variable declarations
	hasMain        bool
	mainHasArgs    bool
	mainParamCount int
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
		emitted: make(map[types.TypeID]bool),
	}
}

// ProcessModule walks all top-level declarations in the AST and collects
// the corresponding C declarations. Call this before EmitTo.
func (e *DeclEmitter) ProcessModule() {
	// Pre-traverse all AST nodes to collect and enqueue all referenced compound types
	g := NewExprGen(e.table, e.intern, e.symbols, e.tree, e.queue)
	visited := make(map[types.TypeID]bool)
	for idx := 0; idx < e.tree.NodeCount(); idx++ {
		tID := g.NodeType(uint32(idx))
		if tID != types.TypeUnknown && tID != 0 {
			gen := isGeneric(e.table, tID, visited)
			if strings.Contains(resolveName(e.table.Entry(tID).NameID, e.intern), "Mutex") || strings.Contains(resolveName(e.table.Entry(tID).NameID, e.intern), "RwLock") || strings.Contains(resolveName(e.table.Entry(tID).NameID, e.intern), "Guard") {
				fmt.Printf("DEBUG Node %d: TypeID %d, Kind %d (%s), isGeneric %t, Name %s\n", idx, tID, e.table.Entry(tID).Kind, fmt.Sprintf("%d", e.table.Entry(tID).Kind), gen, resolveName(e.table.Entry(tID).NameID, e.intern))
			}
			if !gen {
				CTypeName(tID, e.table, e.intern, e.queue)
			}
		}
	}

	root := e.tree.Node(0) // NodeProgram
	child := root.FirstChild
	for child != ast.NullIdx {
		node := e.tree.Node(child)
		switch node.Kind {
		case ast.NodeStructDecl:
			e.processStruct(child, node)
			sChild := node.FirstChild
			for sChild != ast.NullIdx {
				sNode := e.tree.Node(sChild)
				if sNode.Kind == ast.NodeFuncDecl {
					e.processFunc(sChild, sNode)
				}
				sChild = sNode.NextSibling
			}
		case ast.NodeFuncDecl:
			e.processFunc(child, node)
		case ast.NodeConstDecl:
			e.processConst(child, node)
		case ast.NodeVarDecl:
			e.processGlobalVar(child, node)
		case ast.NodeTypeAliasDecl:
			e.processTypeAlias(child, node)
		}
		child = node.NextSibling
	}

	// Drain the type declaration queue for any types referenced
	// by function signatures or global declarations.
	e.drainTypeDecls()

	// Process all instantiated generic functions/methods
	var instKeys []uint32
	for k := range e.symbols.InstantiatedToOriginalName {
		instKeys = append(instKeys, k)
	}
	// Sort to preserve perfect determinism
	sort.Slice(instKeys, func(i, j int) bool {
		return instKeys[i] < instKeys[j]
	})

	for _, instSymIdx := range instKeys {
		sym := e.symbols.SymbolAt(instSymIdx)
		if sym.DeclNode != 0 {
			declNode := e.tree.Node(sym.DeclNode)
			if declNode.Kind == ast.NodeFuncDecl {
				e.processFunc(sym.DeclNode, declNode)
			}
		}
	}

	// Drain type declarations again to catch any types enqueued by instantiated generic functions.
	e.drainTypeDecls()
}

func isGeneric(table *types.TypeTable, id types.TypeID, visited map[types.TypeID]bool) bool {
	if id == types.TypeUnknown || id == 0 {
		return false
	}
	if visited[id] {
		return false
	}
	visited[id] = true
	defer func() { visited[id] = false }()

	entry := table.Entry(id)
	switch entry.Kind {
	case types.KindGeneric:
		return true
	case types.KindPointer:
		return isGeneric(table, table.PointerElem(id), visited)
	case types.KindRef:
		return isGeneric(table, types.TypeID(entry.Extra), visited)
	case types.KindSlice:
		return isGeneric(table, table.SliceElem(id), visited)
	case types.KindArray:
		return isGeneric(table, table.ArrayElem(id), visited)
	case types.KindGenericInst:
		for _, arg := range table.GenericInstArgs(id) {
			if isGeneric(table, arg, visited) {
				return true
			}
		}
	case types.KindFunction:
		fi := table.FuncInfo(id)
		if isGeneric(table, fi.Return, visited) {
			return true
		}
		for _, param := range fi.Params {
			if isGeneric(table, param, visited) {
				return true
			}
		}
	case types.KindStruct:
		si := table.StructInfo(id)
		for _, param := range si.GenericParams {
			if isGeneric(table, types.TypeID(param), visited) {
				return true
			}
		}
		for _, field := range si.Fields {
			if isGeneric(table, field.TypeID, visited) {
				return true
			}
		}
	case types.KindSum:
		sumInfo := table.SumInfo(id)
		for _, param := range sumInfo.GenericParams {
			if isGeneric(table, types.TypeID(param), visited) {
				return true
			}
		}
		for _, variant := range sumInfo.Variants {
			if variant.PayloadType != 0 && isGeneric(table, variant.PayloadType, visited) {
				return true
			}
		}
	case types.KindInterface:
		ii := table.InterfaceInfo(id)
		for _, method := range ii.Methods {
			if isGeneric(table, method.Return, visited) {
				return true
			}
			for _, param := range method.Params {
				if isGeneric(table, param, visited) {
					return true
				}
			}
		}
	}
	return false
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
		if e.mainHasArgs {
			fmt.Fprintln(w, "#define AX_MAIN_WITH_ARGS")
		}
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
	symIdx := node.Payload
	var structTypeID types.TypeID
	if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
		sym := e.symbols.SymbolAt(symIdx)
		if sym.Flags&sema.SymFlagGeneric != 0 {
			return
		}
		structTypeID = types.TypeID(sym.TypeID)
	}

	if structTypeID != 0 {
		e.emitted[structTypeID] = true
	}

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
			
			fTypeID := types.TypeID(0)
			symIdx := childNode.Payload
			if symIdx != 0 && e.symbols != nil && int(symIdx) < len(e.symbols.Symbols) && e.symbols.SymbolAt(symIdx).Kind == sema.SymField {
				fTypeID = types.TypeID(e.symbols.SymbolAt(symIdx).TypeID)
			} else {
				fTypeID = types.TypeID(childNode.Payload)
			}
			
			fEntry := e.table.Entry(fTypeID)
			if fEntry.Kind == types.KindArray {
				elemID := e.table.ArrayElem(fTypeID)
				elemC := CTypeName(elemID, e.table, e.intern, e.queue)
				length := e.table.ArrayLength(fTypeID)
				fields = append(fields, fieldInfo{name: fmt.Sprintf("%s[%d]", fName, length), ctype: elemC})
			} else {
				fCType := CTypeName(fTypeID, e.table, e.intern, e.queue)
				fields = append(fields, fieldInfo{name: fName, ctype: fCType})
			}
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
	symIdx := node.Payload
	if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
		sym := e.symbols.SymbolAt(symIdx)
		if sym.Flags&sema.SymFlagGeneric != 0 {
			return
		}
	}

	nameID := e.nodeNameID(idx)
	name := e.resolveName(nameID, idx)
	if name == "main" {
		e.hasMain = true
	}

	if node.Flags&ast.FlagIsExtern != 0 && isStdLibFunc(name) {
		return
	}

	// Determine return type from the symbol's TypeID
	var retType string
	var paramStrs []string

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

	if name == "main" {
		retType = "ax_i32"
		if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
			sym := e.symbols.SymbolAt(symIdx)
			if sym.TypeID != 0 {
				entry := e.table.Entry(types.TypeID(sym.TypeID))
				if entry.Kind == types.KindFunction {
					fi := e.table.FuncInfo(types.TypeID(sym.TypeID))
					e.mainParamCount = len(fi.Params)
				}
			}
		}
		if len(paramStrs) > 0 {
			e.mainHasArgs = true
		}
	} else if retType == "" {
		retType = "void"
	}

	// Visibility
	visibility := ""
	if name != "main" && node.Flags&ast.FlagIsPub == 0 && node.Flags&ast.FlagIsExtern == 0 {
		visibility = "static "
	}

	mangledName := GetFuncMangledName(symIdx, name, e.table, e.symbols, e.intern)

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

	// Find the initializer expression node
	var initNode uint32
	child := node.FirstChild
	for child != ast.NullIdx {
		childNode := e.tree.Node(child)
		if childNode.Kind != ast.NodeTypeExpr && childNode.Kind != ast.NodeGenericParams {
			initNode = child
			break
		}
		child = childNode.NextSibling
	}

	if initNode != 0 {
		eg := NewExprGen(e.table, e.intern, e.symbols, e.tree, e.queue)
		initValStr := eg.Emit(initNode)
		e.globals = append(e.globals, fmt.Sprintf("const %s %s = %s;", ctype, mangledName, initValStr))
	}
}

// processGlobalVar processes a top-level global mutable variable declaration.
func (e *DeclEmitter) processGlobalVar(idx uint32, node *ast.AstNode) {
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
	e.globals = append(e.globals, fmt.Sprintf("extern %s %s;", ctype, mangledName))

	// Find the initializer expression node
	var initNode uint32
	child := node.FirstChild
	for child != ast.NullIdx {
		childNode := e.tree.Node(child)
		if childNode.Kind != ast.NodeTypeExpr {
			initNode = child
			break
		}
		child = childNode.NextSibling
	}

	if initNode != 0 {
		eg := NewExprGen(e.table, e.intern, e.symbols, e.tree, e.queue)
		initValStr := eg.Emit(initNode)
		e.globals = append(e.globals, fmt.Sprintf("%s %s = %s;", ctype, mangledName, initValStr))
	} else {
		// If there is no initializer, default initialize it to 0
		e.globals = append(e.globals, fmt.Sprintf("%s %s = {0};", ctype, mangledName))
	}
}

// processTypeAlias processes a type alias (sum type) declaration.
func (e *DeclEmitter) processTypeAlias(idx uint32, node *ast.AstNode) {
	// Type aliases that map to sum types need their tag enum + struct emitted
	symIdx := node.Payload
	if symIdx == 0 || int(symIdx) >= len(e.symbols.Symbols) {
		return
	}
	sym := e.symbols.SymbolAt(symIdx)
	if sym.Flags&sema.SymFlagGeneric != 0 {
		return
	}
	if sym.TypeID == 0 {
		return
	}

	typeID := types.TypeID(sym.TypeID)
	e.emitted[typeID] = true

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
	for e.queue.Len() > 0 {
		ids := e.queue.Drain()
		for _, id := range ids {
			entry := e.table.Entry(id)
			fmt.Printf("[DRAIN] ID: %d, Kind: %d, Name: %s, Emitted: %v\n", id, entry.Kind, resolveName(entry.NameID, e.intern), e.emitted[id])
			if e.emitted[id] {
				continue
			}
			switch entry.Kind {
			case types.KindStruct, types.KindSlice, types.KindGenericInst, types.KindSum:
				e.emitted[id] = true
				decl := CTypeDecl(id, e.table, e.intern, e.queue)
				if decl != "" {
					e.typedefs = append(e.typedefs, decl)
					if entry.Kind == types.KindStruct || entry.Kind == types.KindSum {
						structName := "ax_" + resolveName(entry.NameID, e.intern)
						e.forwards = append(e.forwards, fmt.Sprintf("struct %s;", structName))
					} else if entry.Kind == types.KindGenericInst {
						cname := CTypeName(id, e.table, e.intern, e.queue)
						if strings.HasPrefix(cname, "struct ") {
							structName := strings.TrimPrefix(cname, "struct ")
							e.forwards = append(e.forwards, fmt.Sprintf("struct %s;", structName))
						}
					}
				}
			}
		}
	}
}

// buildFuncParams extracts parameter names from AST and types from FuncType.
func (e *DeclEmitter) buildFuncParams(node *ast.AstNode, fi *types.FuncType) []string {
	// Walk AST children to get parameter names and flags
	var paramNames []string
	var paramFlags []uint16
	child := node.FirstChild
	for child != ast.NullIdx {
		childNode := e.tree.Node(child)
		if childNode.Kind == ast.NodeParamDecl {
			pNameID := e.nodeNameID(child)
			paramNames = append(paramNames, e.resolveName(pNameID, child))
			paramFlags = append(paramFlags, childNode.Flags)
		}
		child = childNode.NextSibling
	}

	// Build param strings combining types and names using EmitParamDecl
	params := make([]string, len(fi.Params))
	for i, pt := range fi.Params {
		pname := fmt.Sprintf("p%d", i) // fallback name
		var flags uint16
		if i < len(paramNames) {
			pname = paramNames[i]
			flags = paramFlags[i]
		}
		params[i] = EmitParamDecl(pname, pt, flags, e.table, e.intern, e.queue)
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

func GetFuncMangledName(symIdx uint32, defaultName string, table *types.TypeTable, symbols *sema.SymbolTable, intern *ast.InternPool) string {
	if defaultName == "main" {
		return "ax_main_usr"
	}
	if defaultName == "free" {
		if symIdx != 0 && symbols != nil && int(symIdx) < len(symbols.Symbols) {
			sym := symbols.SymbolAt(symIdx)
			if sym.Flags&sema.SymFlagExtern != 0 {
				return "free"
			}
		}
		return "ax_free"
	}
	if isStdLibFunc(defaultName) {
		return defaultName
	}
	if symIdx == 0 || symbols == nil || int(symIdx) >= len(symbols.Symbols) {
		if strings.HasPrefix(defaultName, "_AX_") {
			return "ax" + defaultName
		}
		return "ax_" + defaultName
	}
	sym := symbols.SymbolAt(symIdx)
	if sym.Flags&sema.SymFlagExtern != 0 {
		return defaultName
	}
	if sym.TypeID != 0 {
		entry := table.Entry(types.TypeID(sym.TypeID))
		if entry.Kind == types.KindFunction {
			fi := table.FuncInfo(types.TypeID(sym.TypeID))
			if len(fi.Params) > 0 {
				if structName, ok := getReceiverStructName(fi.Params[0], table, intern); ok {
					methodName := defaultName
					if symbols.InstantiatedToOriginalName != nil {
						if origNameID, ok := symbols.InstantiatedToOriginalName[symIdx]; ok {
							methodName = intern.Get(origNameID)
						}
					}
					return "ax_" + structName + "_" + methodName
				}
			}
		}
	}
	if sym.NameID != 0 && intern != nil {
		symName := intern.Get(sym.NameID)
		if strings.HasPrefix(symName, "_AX_") {
			return "ax" + symName
		}
	}
	if strings.HasPrefix(defaultName, "_AX_") {
		return "ax" + defaultName
	}
	return "ax_" + defaultName
}

func getReceiverStructName(t1 types.TypeID, table *types.TypeTable, intern *ast.InternPool) (string, bool) {
	if t1 == types.TypeUnknown || t1 == 0 {
		return "", false
	}
	entry := table.Entry(t1)
	if entry.Kind == types.KindPointer {
		return getReceiverStructName(table.PointerElem(t1), table, intern)
	}
	if entry.Kind == types.KindRef {
		return getReceiverStructName(types.TypeID(entry.Extra), table, intern)
	}
	if entry.Kind == types.KindStruct {
		if entry.NameID != 0 {
			name := intern.Get(entry.NameID)
			name = strings.ReplaceAll(name, "[", "_")
			name = strings.ReplaceAll(name, "]", "")
			name = strings.ReplaceAll(name, ",", "_")
			name = strings.ReplaceAll(name, " ", "")
			name = strings.ReplaceAll(name, "*", "ptr")
			return name, true
		}
		return "", false
	}
	return "", false
}

var stdLibFuncs = map[string]bool{
	"system":   true,
	"remove":   true,
	"fopen":    true,
	"fclose":   true,
	"fseek":    true,
	"ftell":    true,
	"rewind":   true,
	"fread":    true,
	"fwrite":   true,
	"printf":   true,
	"puts":     true,
	"putchar":  true,
	"getchar":  true,
	"fgetc":    true,
	"fgets":    true,
	"fputc":    true,
	"fputs":    true,
	"sprintf":  true,
	"snprintf": true,
	"fprintf":  true,
	"fflush":   true,
	"perror":   true,
	"scanf":    true,
	"fscanf":   true,
	"sscanf":   true,
	"malloc":   true,
	"free":     true,
	"realloc":  true,
	"calloc":   true,
	"exit":     true,
	"abort":    true,
	"memset":   true,
	"memcpy":   true,
	"memmove":  true,
	"memcmp":   true,
	"strlen":   true,
	"strcmp":   true,
	"strncmp":  true,
	"strcpy":   true,
	"strncpy":  true,
	"strcat":   true,
	"strncat":  true,
	"strchr":   true,
	"strrchr":  true,
	"strstr":   true,
	"atoi":     true,
	"atol":     true,
	"atof":     true,
	"strtol":   true,
	"strtoul":  true,
	"abs":      true,
	"labs":     true,
	"pow":      true,
	"sqrt":     true,
	"sin":      true,
	"cos":      true,
	"tan":      true,
	"asin":     true,
	"acos":     true,
	"atan":     true,
	"atan2":    true,
	"exp":      true,
	"log":      true,
	"log10":    true,
	"floor":    true,
	"ceil":     true,
	"VirtualAlloc": true,
	"VirtualFree":  true,
	"GetStdHandle": true,
	"WriteFile":    true,
	"ReadFile":     true,
	"CreateFileA":  true,
	"CloseHandle":  true,
	"ExitProcess":  true,
}

func isStdLibFunc(name string) bool {
	return stdLibFuncs[name]
}
