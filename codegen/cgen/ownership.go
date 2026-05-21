package cgen

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// OwnershipMode describes how a variable is owned at a given point.
type OwnershipMode int

const (
	ModeValue    OwnershipMode = iota // local owned value, pass by value
	ModeRef                           // lent reference (const pointer in C)
	ModeMutRef                        // mutable borrow (non-const pointer)
	ModeIsolated                      // isolated heap value (pointer, no external refs)
	ModeSink                          // sink parameter (consumed by callee)
)

// OwnershipContext tracks per-variable ownership modes for a function scope.
// Populated during function parameter processing and updated by moves.
type OwnershipContext struct {
	modes map[uint32]OwnershipMode // nameID → ownership mode
}

// NewOwnershipContext creates an empty ownership context.
func NewOwnershipContext() *OwnershipContext {
	return &OwnershipContext{
		modes: make(map[uint32]OwnershipMode),
	}
}

// SetMode sets the ownership mode for a named variable.
func (oc *OwnershipContext) SetMode(nameID uint32, mode OwnershipMode) {
	oc.modes[nameID] = mode
}

// GetMode returns the ownership mode for a variable. Defaults to ModeValue.
func (oc *OwnershipContext) GetMode(nameID uint32) OwnershipMode {
	if mode, ok := oc.modes[nameID]; ok {
		return mode
	}
	return ModeValue
}

// EmitMovePoison emits debug-mode poisoning for a moved-from variable.
// For value types: memset(&src, 0, sizeof(src));
// For heap refs (AxRef): src = (AxRef){.ptr=NULL, .gen_id=0};
func EmitMovePoison(w *IndentWriter, srcName string, typeID types.TypeID, table *types.TypeTable) {
	entry := table.Entry(typeID)

	w.Line("#if AX_DEBUG")
	if entry.Kind == types.KindRef {
		w.Linef("%s = (AxRef){.ptr=NULL, .gen_id=0};", srcName)
	} else {
		w.Linef("memset(&%s, 0, sizeof(%s));", srcName, srcName)
	}
	w.Line("#endif")
}

// EmitParamDecl generates the C parameter declaration with ownership-aware type.
// lent T → const T* name
// mut lent T → T* name
// Isolated[T] → T* name /* Isolated */
// !T (sink) / value → T name (pass by value)
func EmitParamDecl(
	name string,
	typeID types.TypeID,
	flags uint16,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	ctype := CTypeName(typeID, table, intern, queue)

	isLent := (flags & ast.FlagIsLent) != 0
	isMut := (flags & ast.FlagIsMut) != 0
	isSink := (flags & ast.FlagIsSink) != 0

	if isLent && isMut {
		// Mutable borrow: non-const pointer
		return fmt.Sprintf("%s* %s", ctype, name)
	}
	if isLent {
		// Immutable borrow: const pointer
		return fmt.Sprintf("const %s* %s", ctype, name)
	}
	if isSink {
		// Sink: pass by value (C copy semantics = ownership transfer)
		return fmt.Sprintf("%s %s", ctype, name)
	}
	// Default: pass by value
	return fmt.Sprintf("%s %s", ctype, name)
}

// AdaptArgForParam adapts an argument expression for a parameter with the given ownership mode.
// For lent/mutlent/isolated params: wraps expr in "&(expr)" to pass address.
// For sink/value params: returns expr unchanged (pass by value).
func AdaptArgForParam(expr string, mode OwnershipMode) string {
	switch mode {
	case ModeRef, ModeMutRef, ModeIsolated:
		return "&(" + expr + ")"
	default:
		return expr
	}
}

// FieldAccessOp returns "." or "->" depending on whether the object is accessed
// through a pointer (lent/mutlent/isolated) or by value.
func FieldAccessOp(mode OwnershipMode) string {
	switch mode {
	case ModeRef, ModeMutRef, ModeIsolated:
		return "->"
	default:
		return "."
	}
}

// ParamModeFromFlags determines the OwnershipMode from AST node flags.
func ParamModeFromFlags(flags uint16) OwnershipMode {
	isLent := (flags & ast.FlagIsLent) != 0
	isMut := (flags & ast.FlagIsMut) != 0

	if isLent && isMut {
		return ModeMutRef
	}
	if isLent {
		return ModeRef
	}
	if (flags & ast.FlagIsSink) != 0 {
		return ModeSink
	}
	return ModeValue
}
