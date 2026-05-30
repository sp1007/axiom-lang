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
		if st.LazyResolver != nil {
			lr = st.LazyResolver
		} else {
			lr = NewLazyResolver(st, tt, func(m *ModuleInfo, st *SymbolTable, tt *types.TypeTable) error {
			moduleName := intern.Get(m.NameID)
			if moduleName == "std.string" {
				lenID := intern.Intern([]byte("len"))
				sliceID := intern.Intern([]byte("slice"))
				concatID := intern.Intern([]byte("concat"))
				replaceID := intern.Intern([]byte("replace"))
				startsWithID := intern.Intern([]byte("starts_with"))
				endsWithID := intern.Intern([]byte("ends_with"))
				containsID := intern.Intern([]byte("contains"))
				charCountID := intern.Intern([]byte("char_count"))
				trimID := intern.Intern([]byte("trim"))
				toUpperID := intern.Intern([]byte("to_upper"))
				toLowerID := intern.Intern([]byte("to_lower"))

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

				// Define std.string.starts_with
				// fn starts_with(s: str, prefix: str) -> bool
				startsWithSymIdx, _ := st.Define(startsWithID, SymFunc, 0, 0)
				startsWithTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString, types.TypeString}, types.TypeBool, nil)
				st.SymbolAt(startsWithSymIdx).TypeID = uint32(startsWithTypeID)
				m.Exports[startsWithID] = startsWithSymIdx

				// Define std.string.ends_with
				// fn ends_with(s: str, suffix: str) -> bool
				endsWithSymIdx, _ := st.Define(endsWithID, SymFunc, 0, 0)
				endsWithTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString, types.TypeString}, types.TypeBool, nil)
				st.SymbolAt(endsWithSymIdx).TypeID = uint32(endsWithTypeID)
				m.Exports[endsWithID] = endsWithSymIdx

				// Define std.string.contains
				// fn contains(s: str, sub: str) -> bool
				containsIDSymIdx, _ := st.Define(containsID, SymFunc, 0, 0)
				containsTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString, types.TypeString}, types.TypeBool, nil)
				st.SymbolAt(containsIDSymIdx).TypeID = uint32(containsTypeID)
				m.Exports[containsID] = containsIDSymIdx

				// Define std.string.char_count
				// fn char_count(s: str) -> i64
				charCountSymIdx, _ := st.Define(charCountID, SymFunc, 0, 0)
				charCountTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString}, types.TypeI64, nil)
				st.SymbolAt(charCountSymIdx).TypeID = uint32(charCountTypeID)
				m.Exports[charCountID] = charCountSymIdx

				// Define std.string.trim
				// fn trim(s: str) -> str
				trimSymIdx, _ := st.Define(trimID, SymFunc, 0, 0)
				trimTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString}, types.TypeString, nil)
				st.SymbolAt(trimSymIdx).TypeID = uint32(trimTypeID)
				m.Exports[trimID] = trimSymIdx

				// Define std.string.to_upper
				// fn to_upper(s: str) -> str
				toUpperSymIdx, _ := st.Define(toUpperID, SymFunc, 0, 0)
				toUpperTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString}, types.TypeString, nil)
				st.SymbolAt(toUpperSymIdx).TypeID = uint32(toUpperTypeID)
				m.Exports[toUpperID] = toUpperSymIdx

				// Define std.string.to_lower
				// fn to_lower(s: str) -> str
				toLowerSymIdx, _ := st.Define(toLowerID, SymFunc, 0, 0)
				toLowerTypeID := tt.RegisterFunction([]types.TypeID{types.TypeString}, types.TypeString, nil)
				st.SymbolAt(toLowerSymIdx).TypeID = uint32(toLowerTypeID)
				m.Exports[toLowerID] = toLowerSymIdx
			}
			return nil
		})
		}
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

	// Pass 1: Define all top-level symbols first to support forward references
	root := &nr.ast.Nodes[0]
	child := root.FirstChild
	for child != ast.NullIdx {
		childNode := &nr.ast.Nodes[child]
		if childNode.Kind == ast.NodeFuncDecl || childNode.Kind == ast.NodeStructDecl || childNode.Kind == ast.NodeInterfaceDecl || childNode.Kind == ast.NodeConstDecl || childNode.Kind == ast.NodeTypeAliasDecl {
			nameID := childNode.Payload
			var kind SymKind
			var flags SymFlags
			if childNode.Kind == ast.NodeFuncDecl {
				kind = SymFunc
				if childNode.Flags&uint16(ast.FlagIsPub) != 0 {
					flags |= SymFlagPub
				}
				if childNode.Flags&uint16(ast.FlagIsExtern) != 0 {
					flags |= SymFlagExtern
				}
				if childNode.Flags&uint16(ast.FlagIsAsync) != 0 {
					flags |= SymFlagAsync
				}
			} else if childNode.Kind == ast.NodeStructDecl {
				kind = SymStruct
				if childNode.Flags&uint16(ast.FlagIsPub) != 0 {
					flags |= SymFlagPub
				}
			} else if childNode.Kind == ast.NodeInterfaceDecl {
				kind = SymInterface
				if childNode.Flags&uint16(ast.FlagIsPub) != 0 {
					flags |= SymFlagPub
				}
			} else if childNode.Kind == ast.NodeConstDecl {
				kind = SymConst
				if childNode.Flags&uint16(ast.FlagIsPub) != 0 {
					flags |= SymFlagPub
				}
			} else if childNode.Kind == ast.NodeTypeAliasDecl {
				kind = SymTypeAlias
				if childNode.Flags&uint16(ast.FlagIsPub) != 0 {
					flags |= SymFlagPub
				}
			}
			nr.defineSymbol(nameID, kind, flags, child)
		}
		child = childNode.NextSibling
	}

	// Pass 2: Normal AST Walk to resolve bodies and expressions
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

	case ast.NodeMethodSig:
		nr.symtable.PushScope(ScopeFunction)
		nr.resolveChildren(nodeIdx)
		nr.symtable.PopScope()

	case ast.NodeInterfaceDecl:
		nameID := node.Payload
		symIdx := nr.defineSymbol(nameID, SymInterface, 0, nodeIdx)
		node.Payload = symIdx
		
		// Register in TypeTable early so that generic constraints can reference its TypeID
		typeID := nr.types.RegisterInterface(nameID, nil)
		nr.symtable.SymbolAt(symIdx).TypeID = uint32(typeID)

		nr.symtable.PushScope(ScopeBlock)
		// Define `Self` as an implicit type alias within the interface scope.
		// This allows method signatures like `fn compare(self: Self, other: Self)`.
		selfNameID := nr.intern.Intern([]byte("Self"))
		selfSymIdx := nr.defineSymbol(selfNameID, SymTypeAlias, 0, nodeIdx)
		nr.symtable.SymbolAt(selfSymIdx).TypeID = uint32(typeID)
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
				err := nr.lazy.PreloadModule(nameID)
				if err != nil {
					fmt.Printf("[PRELOAD ERROR] Failed to preload module %s: %v\n", nr.intern.Get(nameID), err)
				}
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
		parentScopeIdx := uint32(0)
		
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
		// Check if the entire path LHS.RHS resolves to a module import first
		if symIdx, ok := nr.tryResolveModulePath(nodeIdx); ok {
			node.Payload = symIdx
			nr.resolved[nodeIdx] = true
			break
		}

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
				var symIdx uint32 = 0
				if lhsNode.Kind == ast.NodeIdent || lhsNode.Kind == ast.NodeFieldExpr {
					if nr.resolved[lhsIdx] {
						symIdx = lhsNode.Payload
					}
				} else if lhsNode.Kind == ast.NodeIndexExpr {
					colIdx := lhsNode.FirstChild
					if colIdx != 0 {
						fmt.Printf("[RESOLVE-DEBUG] lhsNode is NodeIndexExpr, colIdx=%d, resolved=%v, payload=%d\n", colIdx, nr.resolved[colIdx], nr.ast.Nodes[colIdx].Payload)
					}
					if colIdx != 0 && nr.resolved[colIdx] {
						symIdx = nr.ast.Nodes[colIdx].Payload
					}
				}

				if symIdx != 0 && int(symIdx) < len(nr.symtable.Symbols) {
					sym := nr.symtable.SymbolAt(symIdx)
					if sym.Kind == SymModule && nr.lazy != nil {
						fieldNameID := node.Payload
						
						resolvedIdx, diag := nr.lazy.ResolveField(sym.NameID, fieldNameID, diagnostics.Pos{})
						if diag != nil {
							nr.errors = append(nr.errors, *diag)
						} else {
							node.Payload = resolvedIdx
							nr.resolved[nodeIdx] = true
						}
					} else if sym.Kind == SymStruct && nr.lazy != nil {
						fieldNameID := node.Payload
						modNameID := nr.lazy.FindModuleOfSymbol(symIdx)
						if modNameID != 0 {
							resolvedIdx, diag := nr.lazy.ResolveField(modNameID, fieldNameID, diagnostics.Pos{})
							if diag == nil {
								node.Payload = resolvedIdx
								nr.resolved[nodeIdx] = true
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

	case ast.NodeGenericType:
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
	if symIdx, found := nr.symtable.ResolveInScope(nameID, nr.symtable.CurrentScope()); found {
		curr := symIdx
		for curr != 0 {
			sym := nr.symtable.SymbolAt(curr)
			if sym.DeclNode == declNode {
				return curr
			}
			curr = sym.NextOverload
		}
	}
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
	var paramTypeIDs []types.TypeID
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
					constraintNameID := nr.ast.Nodes[constraintChild].Payload
					if symIdx, found := nr.resolveType(constraintNameID); found {
						sym := nr.symtable.SymbolAt(symIdx)
						if sym.Kind == SymInterface {
							constraintID = sym.TypeID
						}
					}
				}
			}

			// Register generic type parameter
			typeID := nr.types.RegisterGenericType(gpNameID)
			if constraintID != 0 {
				nr.types.SetGenericConstraint(typeID, types.TypeID(constraintID))
			}
			paramTypeIDs = append(paramTypeIDs, typeID)
			
			// Define in current scope (which is the function/struct scope)
			symIdx, diag := nr.symtable.Define(gpNameID, SymGenericParam, 0, gpChild)
			if diag != nil {
				nr.errors = append(nr.errors, *diag)
			} else {
				sym := nr.symtable.SymbolAt(symIdx)
				sym.TypeID = uint32(typeID)
				gpNode.Payload = symIdx
			}
			
			params = append(params, types.GenericParam{
				NameID:     gpNameID,
				Constraint: constraintID,
			})
		}
		gpChild = nr.ast.Nodes[gpChild].NextSibling
	}

	tmpl := types.NewGenericTemplate(symID, nodeIdx, params, paramTypeIDs, nr.ast)
	nr.types.RegisterGenericTemplate(tmpl)
}

func (nr *NameResolver) resolveType(nameID uint32) (uint32, bool) {
	name := nr.intern.Get(nameID)
	tblNameID := nr.symtable.intern.InternString(name)

	// Search from top of stack (innermost) to bottom (global)
	stack := nr.symtable.GetStack()
	for i := len(stack) - 1; i >= 0; i-- {
		scopeIdx := stack[i]
		if symIdx, found := nr.symtable.Scopes[scopeIdx].get(tblNameID); found {
			sym := nr.symtable.SymbolAt(symIdx)
			if sym.Kind == SymStruct || sym.Kind == SymInterface || sym.Kind == SymTypeAlias || sym.Kind == SymBuiltinType || sym.Kind == SymGenericParam {
				return symIdx, true
			}
		}
	}

	// Fallback to global scope search in case the scope stack doesn't have the global scope for some reason
	if symIdx, found := nr.symtable.ResolveGlobal(tblNameID); found {
		sym := nr.symtable.SymbolAt(symIdx)
		if sym.Kind == SymStruct || sym.Kind == SymInterface || sym.Kind == SymTypeAlias || sym.Kind == SymBuiltinType || sym.Kind == SymGenericParam {
			return symIdx, true
		}
	}

	return 0, false
}

func (nr *NameResolver) tryResolveModulePath(nodeIdx uint32) (uint32, bool) {
	node := &nr.ast.Nodes[nodeIdx]
	if nr.resolved[nodeIdx] && node.Payload != 0 && int(node.Payload) < len(nr.symtable.Symbols) {
		sym := nr.symtable.SymbolAt(node.Payload)
		if sym.Kind == SymModule {
			return node.Payload, true
		}
	}
	if node.Kind == ast.NodeIdent {
		fullName := nr.intern.Get(node.Payload)
		fullNameID := nr.intern.InternString(fullName)
		if symIdx, found := nr.symtable.Resolve(fullNameID); found {
			sym := nr.symtable.SymbolAt(symIdx)
			if sym.Kind == SymModule {
				return symIdx, true
			}
		}
	}
	if node.Kind == ast.NodeFieldExpr {
		lhsIdx := node.FirstChild
		if lhsIdx == 0 {
			return 0, false
		}
		lhsName, ok := nr.reconstructPath(lhsIdx)
		if !ok {
			return 0, false
		}
		payload := node.Payload
		if nr.resolved[nodeIdx] && payload != 0 && int(payload) < len(nr.symtable.Symbols) {
			payload = nr.symtable.SymbolAt(payload).NameID
		}
		rhsName := nr.intern.Get(payload)
		fullName := lhsName + "." + rhsName
		fullNameID := nr.intern.InternString(fullName)
		if symIdx, found := nr.symtable.Resolve(fullNameID); found {
			sym := nr.symtable.SymbolAt(symIdx)
			if sym.Kind == SymModule {
				return symIdx, true
			}
		}
	}
	return 0, false
}

func (nr *NameResolver) reconstructPath(nodeIdx uint32) (string, bool) {
	node := &nr.ast.Nodes[nodeIdx]
	if nr.resolved[nodeIdx] && node.Payload != 0 && int(node.Payload) < len(nr.symtable.Symbols) {
		sym := nr.symtable.SymbolAt(node.Payload)
		if sym.Kind == SymModule {
			return nr.intern.Get(sym.NameID), true
		}
	}
	if node.Kind == ast.NodeIdent {
		nameID := node.Payload
		if nr.resolved[nodeIdx] && nameID != 0 && int(nameID) < len(nr.symtable.Symbols) {
			nameID = nr.symtable.SymbolAt(nameID).NameID
		}
		return nr.intern.Get(nameID), true
	}
	if node.Kind == ast.NodeFieldExpr {
		lhsIdx := node.FirstChild
		if lhsIdx == 0 {
			return "", false
		}
		lhsName, ok := nr.reconstructPath(lhsIdx)
		if !ok {
			return "", false
		}
		nameID := node.Payload
		if nr.resolved[nodeIdx] && nameID != 0 && int(nameID) < len(nr.symtable.Symbols) {
			nameID = nr.symtable.SymbolAt(nameID).NameID
		}
		rhsName := nr.intern.Get(nameID)
		return lhsName + "." + rhsName, true
	}
	return "", false
}
