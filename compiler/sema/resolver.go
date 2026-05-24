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
	resolved map[uint32]bool
}

// NewNameResolver creates a new NameResolver.
func NewNameResolver(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable, lr *LazyResolver) *NameResolver {
	if lr == nil {
		lr = NewLazyResolver(st, tt, func(m *ModuleInfo, st *SymbolTable, tt *types.TypeTable) error {
			moduleName := intern.Get(m.NameID)
			if moduleName == "std.string" {
				lenID := intern.Intern([]byte("len"))
				sliceID := intern.Intern([]byte("slice"))
				concatID := intern.Intern([]byte("concat"))
				replaceID := intern.Intern([]byte("replace"))

				// Define std.string.len
				// fn len(s: str) -> i64
				lenSymIdx, _ := st.Define(lenID, SymFunc, 0, 0)
				lenTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString}, types.TypeI64, nil)
				st.SymbolAt(lenSymIdx).TypeID = uint32(lenTypeID)
				m.Exports[lenID] = lenSymIdx

				// Define std.string.slice
				// fn slice(s: str, start: i64, end: i64) -> str
				sliceSymIdx, _ := st.Define(sliceID, SymFunc, 0, 0)
				sliceTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString, types.TypeI64, types.TypeI64}, types.TypeString, nil)
				st.SymbolAt(sliceSymIdx).TypeID = uint32(sliceTypeID)
				m.Exports[sliceID] = sliceSymIdx

				// Define std.string.concat
				// fn concat(a: str, b: str) -> str
				concatSymIdx, _ := st.Define(concatID, SymFunc, 0, 0)
				concatTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString, types.TypeString}, types.TypeString, nil)
				st.SymbolAt(concatSymIdx).TypeID = uint32(concatTypeID)
				m.Exports[concatID] = concatSymIdx

				// Define std.string.replace
				// fn replace(s: str, old: str, new: str) -> str
				replaceSymIdx, _ := st.Define(replaceID, SymFunc, 0, 0)
				replaceTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString, types.TypeString, types.TypeString}, types.TypeString, nil)
				st.SymbolAt(replaceSymIdx).TypeID = uint32(replaceTypeID)
				m.Exports[replaceID] = replaceSymIdx
			}
			return nil
		})
	}
	return &NameResolver{
		ast:      tree,
		intern:   intern,
		symtable: st,
		types:    tt,
		lazy:     lr,
		resolved: make(map[uint32]bool),
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
		Pos:      nodePos(nr.ast, nodeIdx),
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
		var flags SymFlags
		if node.Flags&uint16(ast.FlagIsPub) != 0 {
			flags |= SymFlagPub
		}
		if node.Flags&uint16(ast.FlagIsExtern) != 0 {
			flags |= SymFlagExtern
		}
		if node.Flags&uint16(ast.FlagIsAsync) != 0 {
			flags |= SymFlagAsync
		}
		symIdx := nr.defineSymbol(nameID, SymFunc, flags, nodeIdx)
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
		flags := SymFlags(0)
		if node.Flags&uint16(ast.FlagIsMut) != 0 {
			flags |= SymFlagMut
		}
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

		nr.symtable.PushScope(ScopeBlock)

		if node.Flags&uint16(ast.FlagIsGeneric) != 0 {
			nr.registerGenericTemplate(nodeIdx, symIdx)
			sym := nr.symtable.SymbolAt(symIdx)
			sym.Flags |= SymFlagGeneric
		}
		
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

	case ast.NodeIfStmt, ast.NodeElifClause, ast.NodeElseClause, ast.NodeMatchArm, ast.NodeArenaBlock, ast.NodeWhileStmt:
		nr.symtable.PushScope(ScopeBlock)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeForStmt:
		nr.symtable.PushScope(ScopeLoop)
		nameID := node.Payload
		symIdx := nr.defineSymbol(nameID, SymVar, 0, nodeIdx)
		node.Payload = symIdx
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeClosureExpr:
		nr.symtable.PushScope(ScopeClosure)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	// References
	case ast.NodeIdent:
		nameID := node.Payload
		name := nr.intern.Get(nameID)
		if name == "break" || name == "continue" {
			break
		}
		symIdx, found := nr.symtable.Resolve(nameID)
		if name == "free" {
			fmt.Printf("[DEBUG-RESOLVE-FREE] Resolve free: found=%t symIdx=%d\n", found, symIdx)
			for idx, sym := range nr.symtable.Symbols {
				symName := nr.intern.Get(sym.NameID)
				if symName == "free" {
					fmt.Printf("[DEBUG-RESOLVE-FREE]   Symbol idx=%d kind=%v flags=%v scopeID=%d declNode=%d\n", idx, sym.Kind, sym.Flags, sym.ScopeID, sym.DeclNode)
				}
			}
		}
		if !found {
			nr.errorf(nodeIdx, 2010, "undefined: '%s'", name)
		} else {
			node.Payload = symIdx // Modify AST in-place
			nr.symtable.MarkUsed(symIdx)
			nr.resolved[nodeIdx] = true
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
			// Check if LHS is an Ident and LHS.RHS is a module import
			lhsNode := &nr.ast.Nodes[lhsIdx]
			isModuleImport := false
			if lhsNode.Kind == ast.NodeIdent {
				lhsName := nr.intern.Get(lhsNode.Payload)
				rhsName := nr.intern.Get(node.Payload)
				fullName := lhsName + "." + rhsName
				fullNameID := nr.intern.InternString(fullName)
				if symIdx, found := nr.symtable.Resolve(fullNameID); found {
					sym := nr.symtable.SymbolAt(symIdx)
					if sym.Kind == SymModule {
						node.Payload = symIdx
						isModuleImport = true
						nr.resolved[nodeIdx] = true
					}
				}
			}

			if !isModuleImport {
				nr.resolveNode(lhsIdx)
				
				// After resolving LHS, if it's an Ident or FieldExpr that resolved to a SymModule:
				lhsNode = &nr.ast.Nodes[lhsIdx]
				if lhsNode.Kind == ast.NodeIdent || lhsNode.Kind == ast.NodeFieldExpr {
					if nr.resolved[lhsIdx] {
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
									nr.resolved[nodeIdx] = true
								}
							}
						}
					}
				}
			}
		}

	case ast.NodeTypeExpr:
		if node.Payload != 0 {
			nameID := node.Payload
			name := nr.intern.Get(nameID)
			symIdx, found := nr.resolveType(nameID)
			if !found {
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
		if symIdx, found := nr.symtable.Resolve(nameID); found {
			sym := nr.symtable.SymbolAt(symIdx)
			if sym.Kind == SymVariant {
				node.Payload = symIdx
				nr.symtable.MarkUsed(symIdx)
				break
			}
		}
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
		name := nr.intern.Get(nameID)
		diag.Message = fmt.Sprintf("symbol already defined in this scope: '%s'", name)
		diag.Pos = nodePos(nr.ast, declNode)
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

func (nr *NameResolver) resolveType(nameID uint32) (uint32, bool) {
	stack := nr.symtable.GetStack()
	for i := len(stack) - 1; i >= 0; i-- {
		scopeIdx := stack[i]
		if symIdx, found := nr.symtable.Scopes[scopeIdx].get(nameID); found {
			sym := nr.symtable.SymbolAt(symIdx)
			if sym.Kind == SymStruct || sym.Kind == SymInterface || sym.Kind == SymTypeAlias || sym.Kind == SymBuiltinType || sym.Kind == SymGenericParam {
				return symIdx, true
			}
		}
	}
	return 0, false
}
