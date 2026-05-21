package cgen

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// UnsafeBlockGen manages code generation state for unsafe blocks.
// When active, safety checks (bounds checks, generational checks) are suppressed.
type UnsafeBlockGen struct {
	table  *types.TypeTable
	intern *ast.InternPool
	queue  *TypeDeclQueue
}

// NewUnsafeBlockGen creates a new UnsafeBlockGen.
func NewUnsafeBlockGen(table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) *UnsafeBlockGen {
	return &UnsafeBlockGen{
		table:  table,
		intern: intern,
		queue:  queue,
	}
}

// EmitUnsafeBlock wraps the body emission in an unsafe scope.
// It:
//  1. Opens a C block with a /* unsafe */ comment for auditability
//  2. Sets the ExprGen.Unsafe flag to true
//  3. Calls emitBody to generate the block contents
//  4. Restores the ExprGen.Unsafe flag to its previous value
//  5. Closes the block
//
// The emitBody callback receives the IndentWriter for generating statements.
func EmitUnsafeBlock(w *IndentWriter, exprGen *ExprGen, emitBody func()) {
	w.Line("{ /* unsafe */")
	w.Indent()
	oldUnsafe := exprGen.Unsafe
	exprGen.Unsafe = true
	emitBody()
	exprGen.Unsafe = oldUnsafe
	w.Dedent()
	w.Line("}")
}

// UnsafeDeref generates a raw pointer dereference without generational check.
// In unsafe mode, this bypasses ax_deref() and accesses the pointer directly.
//
// Example output: (*((ax_i32*)(ref).ptr))
func UnsafeDeref(ref string, typeID types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	ctype := CTypeName(typeID, table, intern, queue)
	return fmt.Sprintf("(*((%s*)(%s).ptr))", ctype, ref)
}

// UncheckedIndex generates an array/slice index access without bounds checking.
// In unsafe mode, this skips the ax_bounds_check() call.
//
// Example output: (arr).ptr[idx]
func UncheckedIndex(arr string, index string) string {
	return fmt.Sprintf("(%s).ptr[%s]", arr, index)
}

// ArenaBlockGen manages code generation state for arena-scoped blocks.
// It tracks the current arena variable name for redirecting allocations.
type ArenaBlockGen struct {
	table  *types.TypeTable
	intern *ast.InternPool
	queue  *TypeDeclQueue
}

// NewArenaBlockGen creates a new ArenaBlockGen.
func NewArenaBlockGen(table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) *ArenaBlockGen {
	return &ArenaBlockGen{
		table:  table,
		intern: intern,
		queue:  queue,
	}
}

// EmitArenaBlock generates C code for an AXIOM 'in [arena]: block'.
// It wraps the body with arena scope management:
//  1. Opens a C block
//  2. Calls emitBody to generate the block contents
//  3. Emits ax_arena_destroy(arenaVar) after the block closes
//
// The arenaVar is the C variable name holding the AxArena*.
// The emitBody callback generates the block's statements.
func EmitArenaBlock(w *IndentWriter, arenaVar string, emitBody func()) {
	w.Line("{ /* arena block */")
	w.Indent()
	emitBody()
	w.Dedent()
	w.Line("}")
	w.Linef("ax_arena_destroy(%s);", arenaVar)
}

// EmitArenaCreate generates the arena creation statement.
// Example: AxArena* arena_name = ax_arena_create(capacity);
func EmitArenaCreate(w *IndentWriter, varName string, capacity int) {
	w.Linef("AxArena* %s = ax_arena_create(%d);", varName, capacity)
}

// EmitArenaAlloc generates an arena allocation for a variable.
// It emits a declaration and assignment using ax_arena_alloc.
//
// Example:
//
//	struct ax_Node* nodes = (struct ax_Node*)ax_arena_alloc(arena, sizeof(struct ax_Node));
func EmitArenaAlloc(w *IndentWriter, varName string, typeID types.TypeID, arenaVar string, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) {
	ctype := CTypeName(typeID, table, intern, queue)
	w.Linef("%s* %s = (%s*)ax_arena_alloc(%s, sizeof(%s));", ctype, varName, ctype, arenaVar, ctype)
}

// EmitArenaAllocInit generates an arena allocation followed by initialization.
// Example:
//
//	struct ax_Node* n = (struct ax_Node*)ax_arena_alloc(arena, sizeof(struct ax_Node));
//	*n = (struct ax_Node){.x=1, .y=2};
func EmitArenaAllocInit(w *IndentWriter, varName string, typeID types.TypeID, arenaVar string, initExpr string, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) {
	EmitArenaAlloc(w, varName, typeID, arenaVar, table, intern, queue)
	w.Linef("*%s = %s;", varName, initExpr)
}

// EmitArenaReset generates an arena reset call.
// This allows reusing the same arena memory for a new batch of allocations.
// Example: ax_arena_reset(arena);
func EmitArenaReset(w *IndentWriter, arenaVar string) {
	w.Linef("ax_arena_reset(%s);", arenaVar)
}

// EmitArenaDestroy generates an arena destroy call.
// Example: ax_arena_destroy(arena);
func EmitArenaDestroy(w *IndentWriter, arenaVar string) {
	w.Linef("ax_arena_destroy(%s);", arenaVar)
}

// ArenaAllocExpr returns the C expression string for an arena allocation.
// This is used when the allocation is part of a larger expression.
// Example: (struct ax_Node*)ax_arena_alloc(arena, sizeof(struct ax_Node))
func ArenaAllocExpr(typeID types.TypeID, arenaVar string, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	ctype := CTypeName(typeID, table, intern, queue)
	return fmt.Sprintf("(%s*)ax_arena_alloc(%s, sizeof(%s))", ctype, arenaVar, ctype)
}
