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
	emittedProtos map[string]bool

	forwards []string // struct forward declarations
	typedefs []string // struct/sum/slice definitions
	protos   []string // function prototypes
	globals  []string // global variable declarations
	hasMain        bool
	mainHasArgs    bool
	mainParamCount int
	currentModule  string
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
		emittedProtos: make(map[string]bool),
	}
}

// ProcessModule walks all top-level declarations in the AST and collects
// the corresponding C declarations. Call this before EmitTo.
func (e *DeclEmitter) ProcessModule() {
	// 1. Process main tree
	e.currentModule = ""
	e.processTree(e.tree)

	// 2. Process all loaded module trees
	if e.symbols != nil && e.symbols.LazyResolver != nil {
		var modKeys []uint32
		for k := range e.symbols.LazyResolver.GetModules() {
			modKeys = append(modKeys, k)
		}
		sort.Slice(modKeys, func(i, j int) bool {
			return modKeys[i] < modKeys[j]
		})
		for _, k := range modKeys {
			mod := e.symbols.LazyResolver.GetModules()[k]
			if mod.AstTree != nil {
				e.currentModule = string(e.intern.Get(k))
				e.processTree(mod.AstTree)
			}
		}
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
				if sym.TypeID != 0 && isGeneric(e.table, types.TypeID(sym.TypeID), make(map[types.TypeID]bool)) {
					continue
				}
				e.processFunc(sym.DeclNode, declNode)
			}
		}
	}

	// Drain type declarations again to catch any types enqueued by instantiated generic functions.
	e.drainTypeDecls()

	// Scan the AST (main and loaded modules) to find all referenced non-generic, non-extern functions, and generate their prototypes
	referencedFuncs := make(map[uint32]bool)
	visitedSyms := make(map[types.TypeID]bool)
	scanTreeFuncs := func(tree *ast.AstTree) {
		for idx := 0; idx < tree.NodeCount(); idx++ {
			node := tree.Node(uint32(idx))
			symIdx := node.Payload
			if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
				sym := e.symbols.SymbolAt(symIdx)
				if sym.Kind == sema.SymFunc && (sym.Flags&sema.SymFlagGeneric) == 0 && sym.TypeID != 0 {
					if !isGeneric(e.table, types.TypeID(sym.TypeID), visitedSyms) {
						referencedFuncs[symIdx] = true
					}
				}
			}
		}
	}

	scanTreeFuncs(e.tree)
	if e.symbols != nil && e.symbols.LazyResolver != nil {
		for _, mod := range e.symbols.LazyResolver.GetModules() {
			if mod.AstTree != nil {
				scanTreeFuncs(mod.AstTree)
			}
		}
	}

	var symKeys []int
	for symIdx := range referencedFuncs {
		symKeys = append(symKeys, int(symIdx))
	}
	// Sort for perfect determinism
	sort.Ints(symKeys)

	for _, symIdx := range symKeys {
		sym := e.symbols.SymbolAt(uint32(symIdx))
		name := e.intern.Get(sym.NameID)
		if name == "main" {
			continue
		}

		if sym.Flags&sema.SymFlagExtern != 0 {
			if isStdLibFunc(name) {
				continue
			}
		}

		mangledName := GetFuncMangledName(uint32(symIdx), name, e.table, e.symbols, e.intern)
		if e.emittedProtos[mangledName] {
			continue
		}
		e.emittedProtos[mangledName] = true

		entry := e.table.Entry(types.TypeID(sym.TypeID))
		if entry.Kind == types.KindFunction {
			fi := e.table.FuncInfo(types.TypeID(sym.TypeID))
			retType := CTypeName(fi.Return, e.table, e.intern, e.queue)
			if retType == "" {
				retType = "void"
			}

			var paramStrs []string
			for i, pt := range fi.Params {
				pname := fmt.Sprintf("p%d", i)
				paramStrs = append(paramStrs, EmitParamDecl(pname, pt, 0, e.table, e.intern, e.queue))
			}
			if fi.IsVariadic {
				paramStrs = append(paramStrs, "...")
			}

			if len(paramStrs) == 0 {
				e.protos = append(e.protos, fmt.Sprintf("%s %s(void);", retType, mangledName))
			} else {
				e.protos = append(e.protos, fmt.Sprintf("%s %s(%s);",
					retType, mangledName, strings.Join(paramStrs, ", ")))
			}
		}
	}

	// Drain type declarations again to catch any types enqueued by imported function signatures
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

	if node.Flags&ast.FlagIsExtern != 0 {
		if isStdLibFunc(name) {
			return
		}
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
	e.emittedProtos[mangledName] = true

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

	mangledModule := strings.ReplaceAll(e.currentModule, ".", "_")
	mangledName := MangleGlobalName(mangledModule, name)
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

	mangledModule := strings.ReplaceAll(e.currentModule, ".", "_")
	mangledName := MangleGlobalName(mangledModule, name)
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
	var emitTypeWithDependencies func(id types.TypeID)
	emitTypeWithDependencies = func(id types.TypeID) {
		if e.emitted[id] {
			return
		}
		if isGeneric(e.table, id, make(map[types.TypeID]bool)) {
			return
		}
		
		entry := e.table.Entry(id)
		if entry.Kind != types.KindStruct && entry.Kind != types.KindSlice && entry.Kind != types.KindGenericInst && entry.Kind != types.KindSum {
			return
		}

		// Prevent infinite recursion by marking as emitted during processing
		e.emitted[id] = true

		// Emit value dependencies first
		deps := e.getValueDependencies(id)
		for _, dep := range deps {
			emitTypeWithDependencies(dep)
		}

		// Emit this type
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

	for e.queue.Len() > 0 {
		ids := e.queue.Drain()
		for _, id := range ids {
			emitTypeWithDependencies(id)
		}
	}
}

func (e *DeclEmitter) getValueDependencies(id types.TypeID) []types.TypeID {
	var deps []types.TypeID
	entry := e.table.Entry(id)
	
	if entry.Kind == types.KindStruct {
		structInfo := e.table.StructInfo(id)
		for _, f := range structInfo.Fields {
			fBase := e.getBaseType(f.TypeID)
			fEntry := e.table.Entry(fBase)
			if fEntry.Kind == types.KindStruct || fEntry.Kind == types.KindSum || fEntry.Kind == types.KindGenericInst {
				deps = append(deps, fBase)
			}
		}
	} else if entry.Kind == types.KindSum {
		sumInfo := e.table.SumInfo(id)
		for _, v := range sumInfo.Variants {
			if v.PayloadType != 0 && v.PayloadType != types.TypeUnknown {
				fBase := e.getBaseType(v.PayloadType)
				fEntry := e.table.Entry(fBase)
				if fEntry.Kind == types.KindStruct || fEntry.Kind == types.KindSum || fEntry.Kind == types.KindGenericInst {
					deps = append(deps, fBase)
				}
			}
		}
	} else if entry.Kind == types.KindGenericInst {
		// Find base struct/sum template
		var templateID types.TypeID = types.TypeUnknown
		var templateEntry *types.TypeEntry
		for idx := 0; idx < e.table.Count(); idx++ {
			ent := e.table.Entry(types.TypeID(idx))
			if (ent.Kind == types.KindStruct || ent.Kind == types.KindSum) &&
				ent.NameID != 0 && entry.NameID != 0 && ent.NameID == entry.NameID {
				templateID = types.TypeID(idx)
				templateEntry = ent
				break
			}
		}
		if templateEntry != nil {
			typeArgs := e.table.GenericInstArgs(id)
			var genericParams []uint32
			if templateEntry.Kind == types.KindStruct {
				si := e.table.StructInfo(templateID)
				genericParams = si.GenericParams
				for _, f := range si.Fields {
					fType := f.TypeID
					if len(genericParams) > 0 && len(typeArgs) == len(genericParams) {
						fType = e.table.SubstituteGenericType(fType, genericParams, typeArgs)
					}
					fBase := e.getBaseType(fType)
					fEntry := e.table.Entry(fBase)
					if fEntry.Kind == types.KindStruct || fEntry.Kind == types.KindSum || fEntry.Kind == types.KindGenericInst {
						deps = append(deps, fBase)
					}
				}
			} else {
				sumInfo := e.table.SumInfo(templateID)
				genericParams = sumInfo.GenericParams
				for _, v := range sumInfo.Variants {
					if v.PayloadType != 0 && v.PayloadType != types.TypeUnknown {
						fType := v.PayloadType
						if len(genericParams) > 0 && len(typeArgs) == len(genericParams) {
							fType = e.table.SubstituteGenericType(fType, genericParams, typeArgs)
						}
						fBase := e.getBaseType(fType)
						fEntry := e.table.Entry(fBase)
						if fEntry.Kind == types.KindStruct || fEntry.Kind == types.KindSum || fEntry.Kind == types.KindGenericInst {
							deps = append(deps, fBase)
						}
					}
				}
			}
		}
	}
	return deps
}

func (e *DeclEmitter) getBaseType(id types.TypeID) types.TypeID {
	if id == 0 || id == types.TypeUnknown {
		return id
	}
	entry := e.table.Entry(id)
	if entry.Kind == types.KindArray {
		return e.getBaseType(e.table.ArrayElem(id))
	}
	return id
}

// buildFuncParams extracts parameter names from AST and types from FuncType.
func (e *DeclEmitter) buildFuncParams(node *ast.AstNode, fi *types.FuncType) []string {
	// Walk AST children to get parameter names and flags
	var paramNames []string
	var paramFlags []uint16
	var paramIsFuncType []bool
	child := node.FirstChild
	for child != ast.NullIdx {
		childNode := e.tree.Node(child)
		if childNode.Kind == ast.NodeParamDecl {
			pNameID := e.nodeNameID(child)
			paramNames = append(paramNames, e.resolveName(pNameID, child))
			paramFlags = append(paramFlags, childNode.Flags)

			// Check if this parameter has a function type in AST
			isFuncType := false
			c := childNode.FirstChild
			for c != ast.NullIdx {
				cn := e.tree.Node(c)
				if cn.Kind == ast.NodeFuncType {
					isFuncType = true
					break
				}
				c = cn.NextSibling
			}
			paramIsFuncType = append(paramIsFuncType, isFuncType)
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
			if paramIsFuncType[i] {
				pt = e.table.RegisterFunction(nil, types.TypeVoid, nil)
			}
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
	if defaultName == "malloc" {
		if symIdx != 0 && symbols != nil && int(symIdx) < len(symbols.Symbols) {
			sym := symbols.SymbolAt(symIdx)
			if sym.Flags&sema.SymFlagExtern != 0 {
				return "malloc"
			}
		}
		return "ax_alloc"
	}
	if symIdx != 0 && symbols != nil && int(symIdx) < len(symbols.Symbols) {
		sym := symbols.SymbolAt(symIdx)
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
	if defaultName == "starts_with" {
		fmt.Printf("[DEBUG-CGEN-MANGLE] symbolsIsNil=%v lazyIsNil=%v internIsNil=%v symIdx=%d\n", symbols == nil, symbols != nil && symbols.LazyResolver == nil, intern == nil, symIdx)
		if symbols != nil && symbols.LazyResolver != nil && intern != nil {
			fmt.Printf("[DEBUG-CGEN-MANGLE] modulesCount=%d\n", len(symbols.LazyResolver.GetModules()))
			for modNameID, mod := range symbols.LazyResolver.GetModules() {
				fmt.Printf("[DEBUG-CGEN-MANGLE]   mod=%s exportsCount=%d\n", intern.Get(modNameID), len(mod.Exports))
				for nameID, exportedSymIdx := range mod.Exports {
					if intern.Get(nameID) == "starts_with" {
						fmt.Printf("[DEBUG-CGEN-MANGLE]     found starts_with export: exportedSymIdx=%d symIdx=%d\n", exportedSymIdx, symIdx)
					}
				}
			}
		}
	}
	if symbols != nil && symbols.LazyResolver != nil && intern != nil {
		for modNameID, mod := range symbols.LazyResolver.GetModules() {
			for nameID, exportedSymIdx := range mod.Exports {
				if exportedSymIdx == symIdx {
					modName := intern.Get(modNameID)
					fieldName := intern.Get(nameID)
					mangledModule := strings.ReplaceAll(modName, ".", "_")
					return "ax_" + mangledModule + "_" + fieldName
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
	if entry.Kind == types.KindGenericInst {
		typeArgs := table.GenericInstArgs(t1)
		parts := make([]string, len(typeArgs))
		q := NewTypeDeclQueue()
		for i, arg := range typeArgs {
			typeNameStr := CTypeName(arg, table, intern, q)
			parts[i] = sanitizeName(typeNameStr)
		}
		name := resolveName(entry.NameID, intern) + "_" + strings.Join(parts, "_")
		return name, true
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
	"GetLastError": true,
	"GetFileAttributesA": true,
	"CreateDirectoryA": true,
	"RemoveDirectoryA": true,
	"DeleteFileA": true,
	"MoveFileA": true,
	"CopyFileA": true,
	"GetCommandLineW": true,
	"GetCommandLineA": true,
	"GetEnvironmentVariableA": true,
	"SetEnvironmentVariableA": true,
	"ax_str_parse_i64": true,
	"ax_str_parse_f64": true,
	"CreateProcessA":          true,
	"WaitForSingleObject":     true,
	"GetExitCodeProcess":      true,
	"TerminateProcess":        true,
}

func isStdLibFunc(name string) bool {
	return stdLibFuncs[name]
}

func (e *DeclEmitter) processTree(tree *ast.AstTree) {
	oldTree := e.tree
	e.tree = tree
	defer func() { e.tree = oldTree }()

	g := NewExprGen(e.table, e.intern, e.symbols, tree, e.queue)
	visited := make(map[types.TypeID]bool)
	for idx := 0; idx < tree.NodeCount(); idx++ {
		tID := g.NodeType(uint32(idx))
		if tID != types.TypeUnknown && tID != 0 {
			gen := isGeneric(e.table, tID, visited)
			if !gen {
				CTypeName(tID, e.table, e.intern, e.queue)
			}
		}
	}

	root := tree.Node(0) // NodeProgram
	child := root.FirstChild
	for child != ast.NullIdx {
		node := tree.Node(child)
		switch node.Kind {
		case ast.NodeStructDecl:
			if node.Flags&ast.FlagIsGeneric != 0 {
				break
			}
			symIdx := node.Payload
			if symIdx != 0 && int(symIdx) < len(e.symbols.Symbols) {
				sym := e.symbols.SymbolAt(symIdx)
				structTypeID := types.TypeID(sym.TypeID)
				if structTypeID != 0 {
					e.queue.Enqueue(structTypeID)
				}
			}
			sChild := node.FirstChild
			for sChild != ast.NullIdx {
				sNode := tree.Node(sChild)
				if sNode.Kind == ast.NodeFuncDecl {
					if sNode.Flags&ast.FlagIsGeneric == 0 {
						e.processFunc(sChild, sNode)
					}
				}
				sChild = sNode.NextSibling
			}
		case ast.NodeFuncDecl:
			if node.Flags&ast.FlagIsGeneric == 0 {
				e.processFunc(child, node)
			}
		case ast.NodeConstDecl:
			e.processConst(child, node)
		case ast.NodeVarDecl:
			e.processGlobalVar(child, node)
		case ast.NodeTypeAliasDecl:
			e.processTypeAlias(child, node)
		}
		child = node.NextSibling
	}
}
