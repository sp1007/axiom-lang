package sema

import (
	"fmt"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

// SymbolTable is the central symbol storage for the entire compilation unit.
type SymbolTable struct {
	Symbols                    []Symbol // flat array of all symbols across all scopes
	Scopes                     []Scope  // all scopes (index 0 = global scope)
	stack                      []uint32 // active scope stack (indices into Scopes)
	intern                     *ast.InternPool
	InstantiatedToOriginalName map[uint32]uint32
}

// builtinType represents a built-in type to pre-populate in the global scope.
type builtinType struct {
	name   string
	typeID uint32
}

var builtins = []builtinType{
	{"i8", 1}, {"i16", 2}, {"i32", 3}, {"i64", 4},
	{"u8", 5}, {"u16", 6}, {"u32", 7}, {"u64", 8},
	{"f32", 9}, {"f64", 10},
	{"bool", 11}, {"string", 12}, {"char8", 13},
	{"void", 14}, {"isize", 15}, {"usize", 16},
	{"ActorRef", 21}, // types.TypeActorRef
	{"str", 12},      // alias to types.TypeString
	{"ptr", 0},        // placeholder for generic pointer base
	{"null", 0},
	{"alloc", 0},
	{"free", 0},
	{"memcpy", 0},
	{"memset", 0},
	{"print", 0},
	{"println", 0},
	{"compiler_intrinsic", 0},
	{"assert", 0},
	{"panic", 0},
	{"syscall0", 0},
	{"syscall1", 0},
	{"syscall2", 0},
	{"syscall3", 0},
	{"syscall4", 0},
	{"syscall5", 0},
	{"syscall6", 0},
}

// NewSymbolTable creates a SymbolTable with the global scope pre-populated
// with built-in primitive types.
func NewSymbolTable(intern *ast.InternPool) *SymbolTable {
	st := &SymbolTable{
		Symbols:                    make([]Symbol, 0, 1024),
		Scopes:                     make([]Scope, 0, 64),
		stack:                      make([]uint32, 0, 32),
		intern:                     intern,
		InstantiatedToOriginalName: make(map[uint32]uint32),
	}

	// Create global scope (index 0)
	globalScope := Scope{
		Kind:     ScopeGlobal,
		ParentID: 0,
		Depth:    0,
	}
	globalScope.init(32) // enough capacity for builtins
	st.Scopes = append(st.Scopes, globalScope)
	st.stack = append(st.stack, 0) // push global scope

	// Pre-populate built-ins
	for _, b := range builtins {
		nameID := intern.Intern([]byte(b.name))
		// We don't check for errors here because it's a fresh table
		symIdx := uint32(len(st.Symbols))
		st.Symbols = append(st.Symbols, Symbol{
			NameID:   nameID,
			Kind:     SymBuiltinType,
			Flags:    SymFlagPub, // builtins are implicitly public
			TypeID:   b.typeID,
			DeclNode: 0,
			ScopeID:  0,
		})
		st.Scopes[0].put(nameID, symIdx)
	}
	return st
}

// PushScope creates a new child scope and pushes it onto the stack.
// Returns the new scope's index.
func (st *SymbolTable) PushScope(kind ScopeKind) uint32 {
	parentID := st.CurrentScope()
	depth := st.CurrentDepth() + 1
	
	newScope := Scope{
		Kind:     kind,
		ParentID: parentID,
		Depth:    depth,
	}
	newScope.init(8) // small initial capacity
	
	newScopeIdx := uint32(len(st.Scopes))
	st.Scopes = append(st.Scopes, newScope)
	st.stack = append(st.stack, newScopeIdx)
	
	return newScopeIdx
}

// PopScope pops the current scope from the stack.
// Panics if attempting to pop the global scope.
func (st *SymbolTable) PopScope() {
	if len(st.stack) <= 1 {
		panic("cannot pop global scope")
	}
	st.stack = st.stack[:len(st.stack)-1]
}

// CurrentScope returns the index of the innermost active scope.
func (st *SymbolTable) CurrentScope() uint32 {
	return st.stack[len(st.stack)-1]
}

// CurrentDepth returns the current nesting depth.
func (st *SymbolTable) CurrentDepth() uint32 {
	return uint32(len(st.stack) - 1)
}

// GetStack returns a copy of the current scope stack.
func (st *SymbolTable) GetStack() []uint32 {
	stack := make([]uint32, len(st.stack))
	copy(stack, st.stack)
	return stack
}

// SetStack sets the current scope stack to the provided slice.
func (st *SymbolTable) SetStack(stack []uint32) {
	st.stack = stack
}

// Define adds a new symbol to the current scope.
// Returns the symbol index and nil error on success.
// Returns 0 and a Diagnostic if the name is already defined in the current scope.
func (st *SymbolTable) Define(nameID uint32, kind SymKind, flags SymFlags, declNode uint32) (uint32, *diagnostics.Diagnostic) {
	scopeIdx := st.CurrentScope()
	scope := &st.Scopes[scopeIdx]

	// Check for duplicate in current scope
	if prevIdx, found := scope.get(nameID); found {
		// If the previous symbol was a built-in type, allow the new definition to overwrite/shadow it.
		if st.Symbols[prevIdx].Kind == SymBuiltinType {
			symIdx := uint32(len(st.Symbols))
			st.Symbols = append(st.Symbols, Symbol{
				NameID:       nameID,
				Kind:         kind,
				Flags:        flags,
				TypeID:       0, // unresolved initially
				DeclNode:     declNode,
				ScopeID:      scopeIdx,
				NextOverload: 0,
			})
			scope.Overwrite(nameID, symIdx)
			return symIdx, nil
		}

		// If both the previous symbol and the new symbol are functions, allow overloading.
		if st.Symbols[prevIdx].Kind == SymFunc && kind == SymFunc {
			// Traverse the NextOverload chain to find the last overloaded function
			currIdx := prevIdx
			for {
				if st.Symbols[currIdx].NextOverload == 0 {
					break
				}
				currIdx = st.Symbols[currIdx].NextOverload
			}

			symIdx := uint32(len(st.Symbols))
			st.Symbols = append(st.Symbols, Symbol{
				NameID:       nameID,
				Kind:         kind,
				Flags:        flags,
				TypeID:       0, // unresolved initially
				DeclNode:     declNode,
				ScopeID:      scopeIdx,
				NextOverload: 0,
			})
			st.Symbols[currIdx].NextOverload = symIdx
			return symIdx, nil
		}

		diag := &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     2001,
			Message:  "symbol already defined in this scope",
		}
		return 0, diag
	}

	symIdx := uint32(len(st.Symbols))
	st.Symbols = append(st.Symbols, Symbol{
		NameID:       nameID,
		Kind:         kind,
		Flags:        flags,
		TypeID:       0, // unresolved initially
		DeclNode:     declNode,
		ScopeID:      scopeIdx,
		NextOverload: 0,
	})

	scope.put(nameID, symIdx)
	return symIdx, nil
}

// Resolve searches for a symbol from the innermost scope outward.
// Returns the symbol index and true if found, (0, false) if not found.
func (st *SymbolTable) Resolve(nameID uint32) (uint32, bool) {
	// Search from top of stack (innermost) to bottom (global)
	for i := len(st.stack) - 1; i >= 0; i-- {
		scopeIdx := st.stack[i]
		if symIdx, found := st.Scopes[scopeIdx].get(nameID); found {
			return symIdx, true
		}
	}
	return 0, false
}

// ResolveInScope searches for a symbol only in a specific scope.
func (st *SymbolTable) ResolveInScope(nameID uint32, scopeID uint32) (uint32, bool) {
	if int(scopeID) >= len(st.Scopes) {
		return 0, false
	}
	return st.Scopes[scopeID].get(nameID)
}

// ResolveGlobal searches only the global scope (index 0).
func (st *SymbolTable) ResolveGlobal(nameID uint32) (uint32, bool) {
	return st.Scopes[0].get(nameID)
}

// SymbolAt returns a pointer to the symbol at the given index.
// The pointer is invalidated if Symbols slice grows.
func (st *SymbolTable) SymbolAt(idx uint32) *Symbol {
	if int(idx) >= len(st.Symbols) {
		panic(fmt.Sprintf("SymbolAt: index out of bounds: %d >= %d", idx, len(st.Symbols)))
	}
	return &st.Symbols[idx]
}

// MarkMoved sets the SymFlagMoved flag on a symbol.
func (st *SymbolTable) MarkMoved(idx uint32) {
	st.Symbols[idx].Flags |= SymFlagMoved
}

// IsMoved returns true if the symbol has been moved.
func (st *SymbolTable) IsMoved(idx uint32) bool {
	return (st.Symbols[idx].Flags & SymFlagMoved) != 0
}

// MarkUsed sets the SymFlagUsed flag on a symbol.
func (st *SymbolTable) MarkUsed(idx uint32) {
	st.Symbols[idx].Flags |= SymFlagUsed
}
