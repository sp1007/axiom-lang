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
	// cache key: "templateSymID:typeArg1,typeArg2..."
	cache map[string]uint32
}

func NewMonomorphizer(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable) *Monomorphizer {
	return &Monomorphizer{
		ast:      tree,
		intern:   intern,
		symtable: st,
		types:    tt,
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
		panic("Monomorphizer: incorrect number of type arguments")
	}

	// 1. Clone the AST of the generic function
	clonedRoot := m.ast.CloneSubtree(tmpl.NodeIdx)

	// 2. Build substitution map: nameID -> TypeID
	substMap := make(map[uint32]types.TypeID)
	for i, param := range tmpl.Params {
		substMap[param.NameID] = typeArgs[i]
	}

	// 3. Substitute types in the cloned tree
	m.substituteTypeParams(clonedRoot, substMap)

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
	
	nr := NewNameResolver(m.ast, m.intern, m.symtable, m.types, nil)
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

// substituteTypeParams walks the tree and replaces NodeIdent payloads that refer to TypeVars
// with a new Node built-in or type reference.
func (m *Monomorphizer) substituteTypeParams(nodeIdx uint32, subst map[uint32]types.TypeID) {
	node := &m.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeIdent || node.Kind == ast.NodeParamDecl || node.Kind == ast.NodeFieldDecl || node.Kind == ast.NodeFuncDecl || node.Kind == ast.NodeStructDecl || node.Kind == ast.NodeTypeAliasDecl || node.Kind == ast.NodeVariantDecl || node.Kind == ast.NodeTypeExpr || node.Kind == ast.NodeVarDecl || node.Kind == ast.NodeConstDecl || node.Kind == ast.NodeBindingPat || node.Kind == ast.NodeVariantPat || node.Kind == ast.NodeInterfaceDecl || node.Kind == ast.NodeImportDecl || node.Kind == ast.NodeForStmt {
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(m.symtable.Symbols) {
			sym := m.symtable.SymbolAt(symIdx)
			if node.Kind == ast.NodeIdent || node.Kind == ast.NodeTypeExpr || node.Kind == ast.NodeBindingPat || node.Kind == ast.NodeVariantPat {
				if typeID, ok := subst[sym.NameID]; ok {
					entry := m.types.Entry(typeID)
					nameID := entry.NameID
					if nameID == 0 {
						if typeID >= types.TypeI8 && typeID <= types.TypeUSize {
							nameID = m.intern.Intern([]byte(typeID.String()))
						} else {
							nameID = m.intern.Intern([]byte(fmt.Sprintf("type%d", typeID)))
						}
					}
					node.Payload = nameID
				} else {
					node.Payload = sym.NameID
				}
			} else {
				node.Payload = sym.NameID
			}
		}
	}

	child := node.FirstChild
	for child != 0 {
		m.substituteTypeParams(child, subst)
		child = m.ast.Nodes[child].NextSibling
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

func (m *Monomorphizer) mangleName(module, name string, args []types.TypeID) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("_AX_%s_%s", module, name))
	for _, arg := range args {
		entry := m.types.Entry(arg)
		var typeNameStr string
		if entry.NameID == 0 {
			// Primitive types have NameID=0 in TypeTable by default, since they are interned in SymbolTable early.
			// We can infer the name from the TypeID or Kind.
			if arg >= types.TypeI8 && arg <= types.TypeUSize {
				typeNameStr = arg.String()
			} else {
				typeNameStr = fmt.Sprintf("type%d", arg)
			}
		} else {
			typeNameStr = string(m.intern.Get(entry.NameID))
		}
		sb.WriteString(fmt.Sprintf("_%s", typeNameStr))
	}
	return sb.String()
}
