package cgen

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// AllocMode describes how a variable is allocated.
// Determined by escape analysis (Phase 06) and used by the code generator
// to emit the correct allocation, dereference, and cleanup patterns.
type AllocMode int

const (
	AllocStack AllocMode = iota // variable lives on the C stack; no generational check
	AllocHeap                   // variable lives on the heap via ax_alloc; AxRef wrapping
	AllocArena                  // variable lives in an arena; arena-level lifetime
)

// AllocContext tracks per-variable allocation modes for code generation.
// Populated during variable declaration processing.
type AllocContext struct {
	modes map[uint32]AllocMode // variable nameID → allocation mode
}

// NewAllocContext creates an empty allocation context.
func NewAllocContext() *AllocContext {
	return &AllocContext{
		modes: make(map[uint32]AllocMode),
	}
}

// SetMode sets the allocation mode for a variable.
func (ac *AllocContext) SetMode(nameID uint32, mode AllocMode) {
	ac.modes[nameID] = mode
}

// GetMode returns the allocation mode for a variable.
// Default: AllocStack (conservative for correctness: no overhead).
func (ac *AllocContext) GetMode(nameID uint32) AllocMode {
	if mode, ok := ac.modes[nameID]; ok {
		return mode
	}
	return AllocStack
}

// EmitHeapDecl emits a heap-allocated variable declaration.
// Pattern:
//
//	Type* _ax_raw_name = (Type*)ax_alloc(sizeof(Type));
//	*_ax_raw_name = init_expr;
//	AxRef name = ax_make_ref(_ax_raw_name);
func EmitHeapDecl(
	w *IndentWriter,
	name string,
	typeID types.TypeID,
	initExpr string,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) {
	ctype := CTypeName(typeID, table, intern, queue)
	rawName := "_ax_raw_" + name

	w.Linef("%s* %s = (%s*)ax_alloc(sizeof(%s));", ctype, rawName, ctype, ctype)
	if initExpr != "" {
		w.Linef("*%s = %s;", rawName, initExpr)
	}
	w.Linef("AxRef %s = ax_make_ref(%s);", name, rawName)
}

// EmitStackDecl emits a stack-allocated variable declaration.
// Pattern:
//
//	Type name = init_expr;
//	// or: Type name = {0};
func EmitStackDecl(
	w *IndentWriter,
	name string,
	typeID types.TypeID,
	initExpr string,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) {
	ctype := CTypeName(typeID, table, intern, queue)
	if initExpr != "" {
		w.Linef("%s %s = %s;", ctype, name, initExpr)
	} else {
		w.Linef("%s %s = {0};", ctype, name)
	}
}

// EmitArenaDecl emits an arena-allocated variable declaration.
// Pattern:
//
//	Type* name = (Type*)ax_arena_alloc(arena, sizeof(Type));
//	*name = init_expr;
func EmitArenaDecl(
	w *IndentWriter,
	name string,
	typeID types.TypeID,
	initExpr string,
	arenaVar string,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) {
	ctype := CTypeName(typeID, table, intern, queue)
	w.Linef("%s* %s = (%s*)ax_arena_alloc(%s, sizeof(%s));",
		ctype, name, ctype, arenaVar, ctype)
	if initExpr != "" {
		w.Linef("*%s = %s;", name, initExpr)
	}
}

// HeapFieldAccess returns a C expression for accessing a field on a heap-allocated
// variable through the generational reference system.
// Pattern: (((Type*)ax_deref(name))->field)
func HeapFieldAccess(
	varName string,
	fieldName string,
	typeID types.TypeID,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	ctype := CTypeName(typeID, table, intern, queue)
	return fmt.Sprintf("((%s*)ax_deref(%s))->%s", ctype, varName, fieldName)
}

// UnsafeHeapFieldAccess returns a C expression for accessing a field on a heap-allocated
// variable WITHOUT generational checks (used in unsafe blocks).
// Pattern: ((Type*)name.ptr)->field
func UnsafeHeapFieldAccess(
	varName string,
	fieldName string,
	typeID types.TypeID,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	ctype := CTypeName(typeID, table, intern, queue)
	return fmt.Sprintf("((%s*)%s.ptr)->%s", ctype, varName, fieldName)
}

// ArenaFieldAccess returns a C expression for accessing a field on an arena-allocated
// variable (plain pointer, no generational check).
// Pattern: name->field
func ArenaFieldAccess(varName string, fieldName string) string {
	return fmt.Sprintf("%s->%s", varName, fieldName)
}

// StackFieldAccess returns a C expression for accessing a field on a stack-allocated
// variable (direct dot access).
// Pattern: name.field
func StackFieldAccess(varName string, fieldName string) string {
	return fmt.Sprintf("%s.%s", varName, fieldName)
}

// HeapDerefExpr returns a C expression for dereferencing a heap AxRef to get
// the underlying pointer.
// Pattern: ((Type*)ax_deref(ref))
func HeapDerefExpr(
	refExpr string,
	typeID types.TypeID,
	unsafe bool,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	ctype := CTypeName(typeID, table, intern, queue)
	if unsafe {
		return fmt.Sprintf("((%s*)(%s).ptr)", ctype, refExpr)
	}
	return fmt.Sprintf("((%s*)ax_deref(%s))", ctype, refExpr)
}

// FieldAccessForMode returns the correct C expression for field access
// based on the variable's allocation mode.
func FieldAccessForMode(
	varName string,
	fieldName string,
	typeID types.TypeID,
	mode AllocMode,
	unsafe bool,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	switch mode {
	case AllocHeap:
		if unsafe {
			return UnsafeHeapFieldAccess(varName, fieldName, typeID, table, intern, queue)
		}
		return HeapFieldAccess(varName, fieldName, typeID, table, intern, queue)
	case AllocArena:
		return ArenaFieldAccess(varName, fieldName)
	default: // AllocStack
		return StackFieldAccess(varName, fieldName)
	}
}
