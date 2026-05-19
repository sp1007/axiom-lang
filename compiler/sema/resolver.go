package sema

import (
	"fmt"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// NameResolver walks the AST and resolves identifiers to symbols.
type NameResolver struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable
	lazy     *LazyResolver
	errors   []diagnostics.Diagnostic
}

// NewNameResolver creates a new NameResolver.
func NewNameResolver(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable, lr *LazyResolver) *NameResolver {
	return &NameResolver{
		ast:      tree,
		intern:   intern,
		symtable: st,
		types:    tt,
		lazy:     lr,
	}
}

// errorf appends an error diagnostic.
func (nr *NameResolver) errorf(nodeIdx uint32, code int, format string, args ...any) {
	// We lack full pos data in AST nodes currently, just TokenIdx.
	// We'll record the AST node idx as part of the error state or just mock a Pos.
	// For testing, just emitting the message and code is enough.
	nr.errors = append(nr.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		// Mock pos for now
		Pos: diagnostics.Pos{},
	})
}

// Resolve walks the entire AST and resolves all names.
func (nr *NameResolver) Resolve() []diagnostics.Diagnostic {
	if nr.ast == nil || nr.ast.NodeCount() == 0 {
		return nr.errors
	}

	// Start at root
	nr.resolveNode(0)

	// Check unused imports
	if nr.lazy != nil {
		unused := nr.lazy.CheckUnusedImports(nr.intern)
		nr.errors = append(nr.errors, unused...)
	}

	return nr.errors
}

// resolveNode dispatches to specific resolution logic based on NodeKind.
func (nr *NameResolver) resolveNode(nodeIdx uint32) {
	node := &nr.ast.Nodes[nodeIdx]
	
	switch node.Kind {
	case ast.NodeProgram:
		nr.resolveChildren(nodeIdx)

	case ast.NodeFuncDecl:
		// Payload is nameID
		nameID := node.Payload
		symIdx := nr.defineSymbol(nameID, SymFunc, 0, nodeIdx)
		node.Payload = symIdx

		// Push function scope
		nr.symtable.PushScope(ScopeFunction)

		// Register generic template if present
		if node.Flags&uint16(ast.FlagIsGeneric) != 0 {
			nr.registerGenericTemplate(nodeIdx, symIdx)
			// Also mark the symbol as generic
			sym := nr.symtable.SymbolAt(symIdx)
			sym.Flags |= SymFlagGeneric
		}

		// Resolve children (params, ret type, body)
		nr.resolveChildren(nodeIdx)

		nr.symtable.PopScope()

	case ast.NodeParamDecl:
		nameID := node.Payload
		node.Payload = nr.defineSymbol(nameID, SymParam, 0, nodeIdx)
		nr.resolveChildren(nodeIdx)

	case ast.NodeStructDecl:
		nameID := node.Payload
		symIdx := nr.defineSymbol(nameID, SymStruct, 0, nodeIdx)
		node.Payload = symIdx
		nr.symtable.PushScope(ScopeBlock)

		if node.Flags&uint16(ast.FlagIsGeneric) != 0 {
			nr.registerGenericTemplate(nodeIdx, symIdx)
			sym := nr.symtable.SymbolAt(symIdx)
			sym.Flags |= SymFlagGeneric
		}

		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeFieldDecl:
		nameID := node.Payload
		node.Payload = nr.defineSymbol(nameID, SymField, 0, nodeIdx)
		nr.resolveChildren(nodeIdx)

	case ast.NodeInterfaceDecl:
		nameID := node.Payload
		node.Payload = nr.defineSymbol(nameID, SymInterface, 0, nodeIdx)
		nr.symtable.PushScope(ScopeBlock)
		// Define `Self` as an implicit type alias within the interface scope.
		// This allows method signatures like `fn compare(self: Self, other: Self)`.
		selfNameID := nr.intern.Intern([]byte("Self"))
		nr.defineSymbol(selfNameID, SymTypeAlias, 0, nodeIdx)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeVarDecl:
		// VarDecl children: type, init expr. Resolve them first before defining the var
		// so `let x = x` fails as undefined.
		nr.resolveChildren(nodeIdx)
		
		nameID := node.Payload
		// Simplification: check flags for mut
		flags := SymFlags(0) // parse flag parsing not fully integrated here
		node.Payload = nr.defineSymbol(nameID, SymVar, flags, nodeIdx)

	case ast.NodeConstDecl:
		// Const declaration: resolve children (type, init expr), then define symbol.
		nr.resolveChildren(nodeIdx)
		nameID := node.Payload
		node.Payload = nr.defineSymbol(nameID, SymConst, 0, nodeIdx)

	case ast.NodeImportDecl:
		nameID := node.Payload
		if nr.lazy != nil {
			symIdx, diag := nr.lazy.RegisterImport(nameID, "", nodeIdx, nodeIdx)
			if diag != nil {
				nr.errors = append(nr.errors, *diag)
			} else {
				node.Payload = symIdx
			}
		}

	case ast.NodeTypeAliasDecl:
		nameID := node.Payload
		symIdx := nr.defineSymbol(nameID, SymTypeAlias, 0, nodeIdx)
		node.Payload = symIdx

		if node.Flags&uint16(ast.FlagIsGeneric) != 0 {
			nr.registerGenericTemplate(nodeIdx, symIdx)
			sym := nr.symtable.SymbolAt(symIdx)
			sym.Flags |= SymFlagGeneric
		}
		
		nr.symtable.PushScope(ScopeBlock)
		nr.resolveChildren(nodeIdx)
		
		currScopeIdx := nr.symtable.CurrentScope()
		nr.symtable.PopScope()
		parentScopeIdx := nr.symtable.CurrentScope()
		
		currScope := &nr.symtable.Scopes[currScopeIdx]
		parentScope := &nr.symtable.Scopes[parentScopeIdx]
		
		for _, entry := range currScope.entries {
			if entry.nameID == 0 {
				continue
			}
			sym := nr.symtable.SymbolAt(entry.symbolIdx)
			if sym.Kind == SymVariant {
				parentScope.put(entry.nameID, entry.symbolIdx)
			}
		}

	case ast.NodeSumType:
		nr.resolveChildren(nodeIdx)

	case ast.NodeVariantDecl:
		nameID := node.Payload
		symIdx := nr.defineSymbol(nameID, SymVariant, 0, nodeIdx)
		node.Payload = symIdx
		nr.resolveChildren(nodeIdx)

	// Scopes
	case ast.NodeBlock:
		// NodeBlock doesn't push a scope by itself, it's the constructs (IfStmt, etc) that do.
		// Wait, if it's a bare block it might need one. We'll let the parents push scopes.
		nr.resolveChildren(nodeIdx)

	case ast.NodeIfStmt, ast.NodeElifClause, ast.NodeElseClause, ast.NodeMatchArm, ast.NodeArenaBlock:
		nr.symtable.PushScope(ScopeBlock)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeForStmt:
		nr.symtable.PushScope(ScopeLoop)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeClosureExpr:
		nr.symtable.PushScope(ScopeClosure)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	// References
	case ast.NodeIdent:
		nameID := node.Payload
		symIdx, found := nr.symtable.Resolve(nameID)
		if !found {
			name := nr.intern.Get(nameID)
			nr.errorf(nodeIdx, 2010, "undefined: '%s'", name)
		} else {
			node.Payload = symIdx // Modify AST in-place
			nr.symtable.MarkUsed(symIdx)
		}
		// Ident has no children, but call resolveChildren anyway (it's safe)
		nr.resolveChildren(nodeIdx)

	case ast.NodeFieldExpr:
		// FieldExpr is `lhs.rhs`.
		// Resolving this fully requires type checking if `lhs` is a struct, but for lazy module imports,
		// we can resolve it if `lhs` is a SymModule.
		// We'll resolve the LHS first.
		lhsIdx := node.FirstChild
		if lhsIdx != 0 {
			nr.resolveNode(lhsIdx)
			
			// After resolving LHS, if it's an Ident that resolved to a SymModule:
			lhsNode := &nr.ast.Nodes[lhsIdx]
			if lhsNode.Kind == ast.NodeIdent {
				symIdx := lhsNode.Payload
				if symIdx != 0 && int(symIdx) < len(nr.symtable.Symbols) {
					sym := nr.symtable.SymbolAt(symIdx)
					if sym.Kind == SymModule && nr.lazy != nil {
						// RHS is the field name. In AST it's usually stored in Payload of NodeFieldExpr,
						// or as a child NodeIdent. Let's assume Payload has the string ID.
						fieldNameID := node.Payload
						
						resolvedIdx, diag := nr.lazy.ResolveField(sym.NameID, fieldNameID, diagnostics.Pos{})
						if diag != nil {
							nr.errors = append(nr.errors, *diag)
						} else {
							node.Payload = resolvedIdx
						}
					}
				}
			}
		}

	case ast.NodeTypeExpr:
		if node.Payload != 0 {
			nameID := node.Payload
			symIdx, found := nr.symtable.Resolve(nameID)
			if !found {
				name := nr.intern.Get(nameID)
				nr.errorf(nodeIdx, 2010, "undefined type: '%s'", name)
			} else {
				node.Payload = symIdx
				nr.symtable.MarkUsed(symIdx)
			}
		}
		nr.resolveChildren(nodeIdx)

	case ast.NodeBindingPat:
		// A binding pattern introduces a new variable in the match arm scope.
		// e.g., `Some(v)` → `v` is a BindingPat
		nameID := node.Payload
		node.Payload = nr.defineSymbol(nameID, SymVar, 0, nodeIdx)

	case ast.NodeVariantPat:
		// A variant pattern resolves the variant name and processes inner bindings.
		// e.g., `Some(v)` → `Some` is the variant, `v` is a BindingPat child.
		nameID := node.Payload
		symIdx, found := nr.symtable.Resolve(nameID)
		if !found {
			name := nr.intern.Get(nameID)
			nr.errorf(nodeIdx, 2010, "undefined: '%s'", name)
		} else {
			node.Payload = symIdx
			nr.symtable.MarkUsed(symIdx)
		}
		// Resolve child patterns (binding names inside the variant)
		nr.resolveChildren(nodeIdx)

	default:
		nr.resolveChildren(nodeIdx)
	}
}

func (nr *NameResolver) resolveChildren(nodeIdx uint32) {
	node := &nr.ast.Nodes[nodeIdx]
	child := node.FirstChild
	for child != 0 {
		nr.resolveNode(child)
		child = nr.ast.Nodes[child].NextSibling
	}
}

func (nr *NameResolver) defineSymbol(nameID uint32, kind SymKind, flags SymFlags, declNode uint32) uint32 {
	idx, diag := nr.symtable.Define(nameID, kind, flags, declNode)
	if diag != nil {
		nr.errors = append(nr.errors, *diag)
	}
	return idx
}

func (nr *NameResolver) registerGenericTemplate(nodeIdx uint32, symID uint32) {
	node := &nr.ast.Nodes[nodeIdx]
	
	// Find NodeGenericParams child
	var gpNodeIdx uint32
	child := node.FirstChild
	for child != 0 {
		if nr.ast.Nodes[child].Kind == ast.NodeGenericParams {
			gpNodeIdx = child
			break
		}
		child = nr.ast.Nodes[child].NextSibling
	}

	if gpNodeIdx == 0 {
		return
	}

	var params []types.GenericParam
	gpChild := nr.ast.Nodes[gpNodeIdx].FirstChild
	for gpChild != 0 {
		gpNode := &nr.ast.Nodes[gpChild]
		if gpNode.Kind == ast.NodeGenericParam {
			gpNameID := gpNode.Payload
			
			// Constraints
			var constraintID uint32
			constraintChild := gpNode.FirstChild
			if constraintChild != 0 {
				if nr.ast.Nodes[constraintChild].Kind == ast.NodeTypeExpr {
					// We'll leave constraintID = 0 for now since we resolve lazily, 
					// or we can resolve the name to a TypeID if already defined.
					_ = nr.ast.Nodes[constraintChild].Payload
				}
			}

			// Register generic type parameter
			typeID := nr.types.RegisterGenericType(gpNameID)
			
			// Define in current scope (which is the function/struct scope)
			symIdx, diag := nr.symtable.Define(gpNameID, SymGenericParam, 0, gpChild)
			if diag != nil {
				nr.errors = append(nr.errors, *diag)
			} else {
				sym := nr.symtable.SymbolAt(symIdx)
				sym.TypeID = uint32(typeID)
			}
			
			params = append(params, types.GenericParam{
				NameID:     gpNameID,
				Constraint: constraintID,
			})
		}
		gpChild = nr.ast.Nodes[gpChild].NextSibling
	}

	tmpl := types.NewGenericTemplate(symID, nodeIdx, params)
	nr.types.RegisterGenericTemplate(tmpl)
}
