package builder

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p09-t10: Ownership Operations
//
// This file provides helpers for lowering ownership-related operations
// to AIR: heap allocation (OpAlloc + OpMakeRef), dereference with
// generational checks (OpDeref), move semantics (OpMove), and
// arena allocation (OpArenaAlloc).
//
// The funcLowering dispatches to these helpers when it encounters
// ownership-annotated AST nodes (FlagEscapesToHeap, FlagUsesArena,
// FlagIsMoved, etc.) or CTGC-injected nodes (NodeDestroyStmt,
// NodeAliasStmt).
// --------------------------------------------------------------------------

// emitHeapAlloc emits a heap allocation: OpAlloc + OpMakeRef.
// Returns the register holding the generational reference (AxRef).
func (fl *funcLowering) emitHeapAlloc(typeID uint16) uint32 {
	ptrReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpAlloc,
		TypeID: typeID,
		Dest:   ptrReg,
	})

	refReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpMakeRef,
		TypeID: typeID,
		Dest:   refReg,
		Src1:   ptrReg,
	})
	return refReg
}

// emitDeref emits a generational dereference check: OpDeref.
// This validates the reference's gen_id matches the header before access.
// Returns the register holding the raw pointer.
func (fl *funcLowering) emitDeref(refReg uint32, typeID uint16) uint32 {
	ptrReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpDeref,
		TypeID: typeID,
		Dest:   ptrReg,
		Src1:   refReg,
	})
	return ptrReg
}

// emitMove emits an ownership move: OpMove.
// After a move, the source register is poisoned and must not be used.
func (fl *funcLowering) emitMove(srcReg uint32, typeID uint16) uint32 {
	destReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpMove,
		TypeID: typeID,
		Dest:   destReg,
		Src1:   srcReg,
	})
	return destReg
}

// emitFree emits OpFree to release a heap allocation.
func (fl *funcLowering) emitFree(ptrReg uint32) {
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpFree,
		Src1:   ptrReg,
	})
}

// emitArenaAlloc emits an arena allocation: OpArenaAlloc.
// Arena allocations do NOT use OpMakeRef (no generational tracking).
func (fl *funcLowering) emitArenaAlloc(arenaReg uint32, typeID uint16) uint32 {
	ptrReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpArenaAlloc,
		TypeID: typeID,
		Dest:   ptrReg,
		Src1:   arenaReg,
	})
	return ptrReg
}

// emitAliasReuse emits an alias reuse operation: OpAliasReuse.
// This is a CTGC optimization where a destroy + alloc of the same type
// can be combined to reuse the same memory region.
func (fl *funcLowering) emitAliasReuse(ptrReg uint32, typeID uint16) uint32 {
	destReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpAliasReuse,
		TypeID: typeID,
		Dest:   destReg,
		Src1:   ptrReg,
	})
	return destReg
}

// lowerOwnershipAware lowers a VarDecl with ownership awareness.
// If the variable escapes to heap, it emits heap allocation + make_ref.
// If it uses an arena, it emits arena allocation.
// Otherwise, it falls through to stack allocation (default SSA).
func (fl *funcLowering) lowerOwnershipAware(idx uint32, node *ast.AstNode, initReg uint32) uint32 {
	if node.Flags&ast.FlagEscapesToHeap != 0 {
		// Heap allocation
		typeID := uint16(0)
		if node.ExtraIdx != 0 {
			typeID = uint16(node.ExtraIdx)
		}
		refReg := fl.emitHeapAlloc(typeID)

		// Store the initial value into the heap slot
		if initReg != 0 {
			ptrReg := fl.emitDeref(refReg, typeID)
			fl.fb.Emit(air.AirInst{
				Opcode: air.OpStore,
				Src1:   initReg,
				Src2:   ptrReg,
			})
		}
		return refReg
	}

	if node.Flags&ast.FlagUsesArena != 0 {
		typeID := uint16(0)
		if node.ExtraIdx != 0 {
			typeID = uint16(node.ExtraIdx)
		}
		// Arena allocation (arena reg = 0 for MVP — uses implicit current arena)
		ptrReg := fl.emitArenaAlloc(0, typeID)
		if initReg != 0 {
			fl.fb.Emit(air.AirInst{
				Opcode: air.OpStore,
				Src1:   initReg,
				Src2:   ptrReg,
			})
		}
		return ptrReg
	}

	if node.Flags&ast.FlagIsMoved != 0 && initReg != 0 {
		typeID := uint16(0)
		if node.ExtraIdx != 0 {
			typeID = uint16(node.ExtraIdx)
		}
		return fl.emitMove(initReg, typeID)
	}

	// Default: stack allocation (just return the SSA register)
	return initReg
}

// lowerAliasStmt lowers a compiler-injected alias reuse statement (CTGC).
func (fl *funcLowering) lowerAliasStmt(idx uint32, node *ast.AstNode) {
	if node.FirstChild != ast.NullIdx {
		cn := fl.mb.tree.Node(node.FirstChild)
		srcReg := fl.lowerExpr(node.FirstChild, cn)

		typeID := uint16(0)
		if node.ExtraIdx != 0 {
			typeID = uint16(node.ExtraIdx)
		}
		fl.emitAliasReuse(srcReg, typeID)
	}
}
