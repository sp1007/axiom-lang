package sema

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// Monomorphizer is responsible for cloning and specializing generic functions and structs.
type Monomorphizer struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable
	ParentNodeTypes map[uint32]types.TypeID
	// cache key: "templateSymID:typeArg1,typeArg2..."
	cache map[string]uint32
}

func NewMonomorphizer(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable) *Monomorphizer {
	return &Monomorphizer{
		ast:      tree,
		intern:   intern,
		symtable: st,
		types:    tt,
		ParentNodeTypes: nil,
		cache:    make(map[string]uint32),
	}
}

// InstantiateFunction monomorphizes a generic function with concrete type arguments.
// Returns the SymID of the instantiated function.
func (m *Monomorphizer) InstantiateFunction(templateSymID uint32, typeArgs []types.TypeID) (uint32, []diagnostics.Diagnostic) {
	key := m.makeCacheKey(templateSymID, typeArgs)
	if instSymID, ok := m.cache[key]; ok {
		return instSymID, nil
	}

	tmpl, found := m.types.FindGenericTemplate(templateSymID)
	if !found {
		panic("Monomorphizer: generic template not found")
	}

	if len(tmpl.Params) != len(typeArgs) {
		tmplName := "unknown"
		if tmpl.NodeIdx != 0 {
			sym := m.symtable.SymbolAt(templateSymID)
			tmplName = string(m.intern.Get(sym.NameID))
		}
		var paramTypes []uint32
		for _, p := range tmpl.Params {
			paramTypes = append(paramTypes, uint32(p.NameID))
		}
		var argTypes []uint32
		for _, a := range typeArgs {
			argTypes = append(argTypes, uint32(a))
		}
		fmt.Printf("[DEBUG-ARITY-FAIL] templateSymID=%d name=%s tmpl.Params=%v typeArgs=%v\n", templateSymID, tmplName, paramTypes, argTypes)
		panic("Monomorphizer: incorrect number of type arguments")
	}

	// Find source tree for the template symbol
	srcTree := m.ast
	if tmpl.SrcTree != nil {
		srcTree = tmpl.SrcTree
	}

	// 1. Clone the AST of the generic function from the source tree
	clonedRoot := m.ast.CloneSubtreeFrom(srcTree, tmpl.NodeIdx)

	// 2. Build substitution map: nameID -> TypeID
	substMap := make(map[uint32]types.TypeID)
	for i, param := range tmpl.Params {
		substMap[param.NameID] = typeArgs[i]
	}

	// 3. Substitute types in the cloned tree
	m.substituteTypeParams(clonedRoot, tmpl.NodeIdx, srcTree, substMap)

	// 4. Remove generic flag
	m.ast.ClearFlags(clonedRoot, ast.FlagIsGeneric)

	// 5. Remove GenericParams node from children
	m.removeGenericParamsChild(clonedRoot)

	// 6. Mangle the function name
	origSym := m.symtable.SymbolAt(templateSymID)
	origName := m.intern.Get(origSym.NameID)
	mangledName := m.mangleName("std", string(origName), typeArgs)
	mangledNameID := m.intern.Intern([]byte(mangledName))

	// Update payload of cloned func decl
	m.ast.SetPayload(clonedRoot, mangledNameID)

	// 7. Re-run name resolution and type checking on the clone in the GLOBAL scope context
	savedStack := m.symtable.GetStack()
	m.symtable.SetStack([]uint32{0}) // Force global scope
	
	nr := NewNameResolver(m.ast, m.intern, m.symtable, m.types, m.symtable.LazyResolver)
	nr.resolveNode(clonedRoot)
	diags := nr.errors

	// Cache the result EARLY to support recursive generic functions
	instSymID := m.ast.Nodes[clonedRoot].Payload
	m.cache[key] = instSymID

	if m.symtable.InstantiatedToOriginalName == nil {
		m.symtable.InstantiatedToOriginalName = make(map[uint32]uint32)
	}
	m.symtable.InstantiatedToOriginalName[instSymID] = origSym.NameID

	if len(diags) == 0 {
		// 1. Pre-register all cloned structs in the subtree
		var preRegisterStructs func(nodeIdx uint32)
		preRegisterStructs = func(nodeIdx uint32) {
			node := &m.ast.Nodes[nodeIdx]
			if node.Kind == ast.NodeStructDecl {
				symIdx := node.Payload
				if symIdx != 0 && int(symIdx) < len(m.symtable.Symbols) {
					sym := m.symtable.SymbolAt(symIdx)
					if sym.TypeID == 0 || sym.TypeID == uint32(types.TypeUnknown) {
						nameID := sym.NameID
						typeID := m.types.RegisterStruct(nameID, nil, nil)
						sym.TypeID = uint32(typeID)
					}
				}
			}
			child := node.FirstChild
			for child != 0 {
				preRegisterStructs(child)
				child = m.ast.Nodes[child].NextSibling
			}
		}
		preRegisterStructs(clonedRoot)

		// 2. Create the InferenceEngine and pre-infer all cloned structs
		ie := NewInferenceEngine(m.ast, m.symtable, m.types, m)
		var preInferStructs func(nodeIdx uint32)
		preInferStructs = func(nodeIdx uint32) {
			node := &m.ast.Nodes[nodeIdx]
			if node.Kind == ast.NodeStructDecl {
				ie.inferNode(nodeIdx, types.TypeUnknown)
			}
			child := node.FirstChild
			for child != 0 {
				preInferStructs(child)
				child = m.ast.Nodes[child].NextSibling
			}
		}
		preInferStructs(clonedRoot)

		// 3. Infer the instantiated root node
		ie.inferNode(clonedRoot, types.TypeUnknown)
		diags = append(diags, ie.errors...)

		if len(diags) == 0 {
			tc := NewTypeChecker(m.ast, m.intern, m.symtable, m.types, ie)
			tc.CheckNode(clonedRoot)
			diags = append(diags, tc.errors...)
		}
	}

	// Restore scope stack
	m.symtable.SetStack(savedStack)

	return instSymID, diags
}

func isLiteralKind(kind ast.NodeKind) bool {
	return kind == ast.NodeIntLit ||
		kind == ast.NodeFloatLit ||
		kind == ast.NodeStringLit ||
		kind == ast.NodeCharLit ||
		kind == ast.NodeBoolLit ||
		kind == ast.NodeNilLit
}

// substituteTypeParams walks the tree and replaces NodeIdent payloads that refer to TypeVars
// with a new Node built-in or type reference.
func (m *Monomorphizer) substituteTypeParams(nodeIdx, origNodeIdx uint32, srcTree *ast.AstTree, subst map[uint32]types.TypeID) {
	node := &m.ast.Nodes[nodeIdx]
	origNode := &srcTree.Nodes[origNodeIdx]

	if node.Kind == ast.NodeIdent || node.Kind == ast.NodeParamDecl || node.Kind == ast.NodeFieldDecl || node.Kind == ast.NodeFuncDecl || node.Kind == ast.NodeStructDecl || node.Kind == ast.NodeTypeAliasDecl || node.Kind == ast.NodeVariantDecl || node.Kind == ast.NodeTypeExpr || node.Kind == ast.NodeVarDecl || node.Kind == ast.NodeConstDecl || node.Kind == ast.NodeBindingPat || node.Kind == ast.NodeVariantPat || node.Kind == ast.NodeInterfaceDecl || node.Kind == ast.NodeImportDecl || node.Kind == ast.NodeForStmt {
		var nameID uint32

		hasSymPayload := node.Kind == ast.NodeFuncDecl ||
			node.Kind == ast.NodeParamDecl ||
			node.Kind == ast.NodeStructDecl ||
			node.Kind == ast.NodeFieldDecl ||
			node.Kind == ast.NodeInterfaceDecl ||
			node.Kind == ast.NodeVarDecl ||
			node.Kind == ast.NodeConstDecl ||
			node.Kind == ast.NodeImportDecl ||
			node.Kind == ast.NodeTypeAliasDecl ||
			node.Kind == ast.NodeVariantDecl ||
			node.Kind == ast.NodeForStmt ||
			node.Kind == ast.NodeIdent ||
			node.Kind == ast.NodeBindingPat ||
			node.Kind == ast.NodeVariantPat

		if hasSymPayload {
			symIdx := origNode.Payload
			if symIdx != 0 && int(symIdx) < len(m.symtable.Symbols) {
				nameID = m.symtable.SymbolAt(symIdx).NameID
			}
		}

		if nameID == 0 {
			text := srcTree.NodeText(origNodeIdx)
			nameID = m.intern.Intern(text)
		}
		if nameID != 0 {
			if node.Kind == ast.NodeIdent || node.Kind == ast.NodeTypeExpr || node.Kind == ast.NodeBindingPat || node.Kind == ast.NodeVariantPat {
				if typeID, ok := subst[nameID]; ok {
					entry := m.types.Entry(typeID)
					targetNameID := entry.NameID
					if targetNameID == 0 {
						if typeID >= types.TypeI8 && typeID <= types.TypeUSize {
							tStr := typeID.String()
							if typeID == types.TypeString {
								tStr = "str"
							} else if typeID == types.TypeChar8 {
								tStr = "char"
							}
							targetNameID = m.intern.Intern([]byte(tStr))
						} else {
							targetNameID = m.intern.Intern([]byte(fmt.Sprintf("type%d", typeID)))
						}
					}
					node.Payload = targetNameID
				} else {
					node.Payload = nameID
				}
			} else {
				node.Payload = nameID
			}
		}
	} else if node.Kind == ast.NodeFieldExpr {
		node.Payload = origNode.Payload
	} else if !isLiteralKind(node.Kind) {
		node.Payload = 0
	}

	child := node.FirstChild
	origChild := origNode.FirstChild
	for child != 0 && origChild != 0 {
		m.substituteTypeParams(child, origChild, srcTree, subst)
		child = m.ast.Nodes[child].NextSibling
		origChild = srcTree.Nodes[origChild].NextSibling
	}
}

// removeGenericParamsChild removes the NodeGenericParams child from a NodeFuncDecl/NodeStructDecl.
func (m *Monomorphizer) removeGenericParamsChild(nodeIdx uint32) {
	node := &m.ast.Nodes[nodeIdx]
	
	var prev uint32 = 0
	curr := node.FirstChild
	for curr != 0 {
		if m.ast.Nodes[curr].Kind == ast.NodeGenericParams {
			next := m.ast.Nodes[curr].NextSibling
			if prev == 0 {
				node.FirstChild = next
			} else {
				m.ast.Nodes[prev].NextSibling = next
			}
			break
		}
		prev = curr
		curr = m.ast.Nodes[curr].NextSibling
	}
}

func (m *Monomorphizer) makeCacheKey(symID uint32, args []types.TypeID) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d:", symID))
	for i, arg := range args {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", arg))
	}
	return sb.String()
}

func (m *Monomorphizer) getTypeNameRecursive(arg types.TypeID) string {
	entry := m.types.Entry(arg)
	if entry.Kind == types.KindGenericInst {
		base := string(m.intern.Get(entry.NameID))
		typeArgs := m.types.GenericInstArgs(arg)
		parts := make([]string, len(typeArgs))
		for i, a := range typeArgs {
			parts[i] = m.getTypeNameRecursive(a)
		}
		return base + "__" + strings.Join(parts, "__")
	}
	if entry.NameID != 0 {
		return string(m.intern.Get(entry.NameID))
	}
	if arg >= types.TypeI8 && arg <= types.TypeUSize {
		if arg == types.TypeString {
			return "string"
		}
		return arg.String()
	}
	switch entry.Kind {
	case types.KindPointer:
		return "ptr_" + m.getTypeNameRecursive(types.TypeID(entry.Extra))
	case types.KindRef:
		return "ref_" + m.getTypeNameRecursive(types.TypeID(entry.Extra))
	case types.KindSlice:
		return "slice_" + m.getTypeNameRecursive(types.TypeID(entry.Extra))
	case types.KindArray:
		elem := m.types.ArrayElem(arg)
		length := m.types.ArrayLength(arg)
		return fmt.Sprintf("arr_%d_%s", length, m.getTypeNameRecursive(elem))
	case types.KindFunction:
		fInfo := m.types.FuncInfo(arg)
		resStr := "func_" + m.getTypeNameRecursive(fInfo.Return) + "_args"
		for _, pVal := range fInfo.Params {
			resStr += "_" + m.getTypeNameRecursive(pVal)
		}
		return resStr
	}
	return "type"
}

func (m *Monomorphizer) mangleName(module, name string, args []types.TypeID) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("_AX_%s_%s", module, name))
	for _, arg := range args {
		sb.WriteString("__" + m.getTypeNameRecursive(arg))
	}
	return sb.String()
}
