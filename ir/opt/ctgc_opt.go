package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t05: CTGC (Compile-Time GC) Optimization at AIR Level
//
// Performs three memory optimizations:
// 1. Free+Alloc reuse → OpAliasReuse (same type, no intervening use)
// 2. Redundant gen_id check elimination (OpDeref after immediate OpMakeRef)
// 3. Region allocation grouping (future: multiple same-type allocs → bulk)
// --------------------------------------------------------------------------

// CTGCOptPass implements OptPass for compile-time garbage collection optimizations.
type CTGCOptPass struct{}

func (p *CTGCOptPass) Name() string { return "ctgc" }

// Run applies CTGC optimizations to all functions.
func (p *CTGCOptPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		if eliminateFreeAllocPairs(fn) {
			changed = true
		}
		if eliminateRedundantDeref(fn) {
			changed = true
		}
		if promoteStackAllocationAndInsertFrees(fn) {
			changed = true
		}
	}
	return changed
}

// eliminateFreeAllocPairs finds OpFree followed by OpAlloc of the same type
// with no intervening use of the freed register, and replaces with OpAliasReuse.
func eliminateFreeAllocPairs(fn *air.AirFunc) bool {
	changed := false

	for i := 0; i < len(fn.Insts); i++ {
		freeInst := &fn.Insts[i]
		if freeInst.Opcode != air.OpFree {
			continue
		}

		freedReg := freeInst.Src1
		if freedReg == 0 {
			continue
		}

		// Scan forward for a matching OpAlloc
		for j := i + 1; j < len(fn.Insts); j++ {
			inst := &fn.Insts[j]

			// Skip NOPs
			if inst.Opcode == air.OpNop {
				continue
			}

			// If the freed register is used before the alloc, abort
			if usesReg(inst, freedReg) {
				break
			}

			// If we hit a control flow instruction, abort (can't see across blocks)
			if inst.Opcode.IsTerminator() {
				break
			}

			// Found a matching alloc
			if inst.Opcode == air.OpAlloc {
				// Replace: NOP the free, convert alloc to alias_reuse
				allocDest := inst.Dest
				allocType := inst.TypeID

				// Convert the alloc to AliasReuse
				fn.Insts[j] = air.AirInst{
					Opcode: air.OpAliasReuse,
					TypeID: allocType,
					Dest:   allocDest,
					Src1:   freedReg,
				}

				// NOP the free
				fn.Insts[i] = air.AirInst{Opcode: air.OpNop}

				changed = true
				break
			}
		}
	}
	return changed
}

// eliminateRedundantDeref removes OpDeref instructions that immediately follow
// OpMakeRef in the same basic block, where the reference was just created
// and cannot have escaped.
func eliminateRedundantDeref(fn *air.AirFunc) bool {
	changed := false

	// Track which registers hold freshly-created refs (from OpMakeRef)
	freshRefs := make(map[uint32]bool)

	for i := range fn.Insts {
		inst := &fn.Insts[i]

		// On control flow boundaries, clear tracking
		if inst.Opcode.IsTerminator() || inst.Opcode == air.OpJump ||
			inst.Opcode == air.OpBranch {
			freshRefs = make(map[uint32]bool)
			continue
		}

		// Track fresh refs
		if inst.Opcode == air.OpMakeRef && inst.Dest != 0 {
			freshRefs[inst.Dest] = true
			continue
		}

		// If a fresh ref is used in OpDeref, the gen_id check is redundant
		if inst.Opcode == air.OpDeref && freshRefs[inst.Src1] {
			// Replace with a simple copy (skip the gen_id check)
			inst.Opcode = air.OpCopy
			changed = true
			continue
		}

		// If any instruction stores/sends/calls with a fresh ref,
		// it may escape — invalidate tracking
		if inst.Opcode == air.OpStore || inst.Opcode == air.OpCall ||
			inst.Opcode == air.OpSend || inst.Opcode == air.OpSpawn {
			// Clear all tracking (conservative)
			freshRefs = make(map[uint32]bool)
		}
	}

	return changed
}

// usesReg checks if an instruction uses `reg` as a source operand.
func usesReg(inst *air.AirInst, reg uint32) bool {
	if inst.Src1 == reg {
		return true
	}
	if inst.Src2 == reg && !inst.Opcode.IsControl() {
		return true
	}
	if inst.Dest == reg && (inst.Opcode == air.OpStore || inst.Opcode == air.OpSetField) {
		return true
	}
	return false
}

// promoteStackAllocationAndInsertFrees determines which allocations do not escape,
// and inserts explicit OpFree instructions right before function returns to avoid leaks.
func promoteStackAllocationAndInsertFrees(fn *air.AirFunc) bool {
	if len(fn.Insts) == 0 {
		return false
	}

	// 1. Gather all OpAlloc registers
	allocRegs := make(map[uint32]uint16) // map of reg -> typeID
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpAlloc && inst.Dest != 0 {
			allocRegs[inst.Dest] = inst.TypeID
		}
	}

	if len(allocRegs) == 0 {
		return false
	}

	// 2. Build Disjoint Set Union (DSU) to group alias and copy relationships.
	dsuParent := make(map[uint32]uint32)
	var find func(uint32) uint32
	find = func(x uint32) uint32 {
		if _, exists := dsuParent[x]; !exists {
			dsuParent[x] = x
			return x
		}
		if dsuParent[x] == x {
			return x
		}
		dsuParent[x] = find(dsuParent[x])
		return dsuParent[x]
	}
	union := func(x, y uint32) {
		rx := find(x)
		ry := find(y)
		if rx != ry {
			dsuParent[rx] = ry
		}
	}

	// Connect alias/copy instructions
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}
		switch inst.Opcode {
		case air.OpCopy, air.OpMove, air.OpMakeRef, air.OpGEP, air.OpGetField:
			if inst.Dest != 0 && inst.Src1 != 0 {
				union(inst.Dest, inst.Src1)
			}
		}
	}

	// 3. Identify all escaping registers.
	escapedReps := make(map[uint32]bool)

	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}

		switch inst.Opcode {
		case air.OpReturn:
			if inst.Src1 != 0 {
				escapedReps[find(inst.Src1)] = true
			}
		case air.OpCall:
			// The target callee register/pointer (Src1) escapes if it's a register
			if inst.Src1 != 0 {
				escapedReps[find(inst.Src1)] = true
			}
			// All function arguments in Extras escape
			argStart := inst.Src2
			argCount := uint32(0)
			if argStart < uint32(len(fn.Extras)) {
				argCount = fn.Extras[argStart]
			}
			for idx := uint32(0); idx < argCount; idx++ {
				argRegIdx := argStart + 1 + idx
				if argRegIdx < uint32(len(fn.Extras)) {
					argReg := fn.Extras[argRegIdx]
					if argReg != 0 {
						escapedReps[find(argReg)] = true
					}
				}
			}
		case air.OpStore:
			// If we store the pointer Src1 into another memory location, it escapes
			if inst.Src1 != 0 {
				escapedReps[find(inst.Src1)] = true
			}
		case air.OpSetField:
			// Src1.field[Src2] = Dest
			// The value being stored (Dest) escapes
			if inst.Dest != 0 {
				escapedReps[find(inst.Dest)] = true
			}
		case air.OpSend, air.OpSpawn:
			if inst.Src1 != 0 {
				escapedReps[find(inst.Src1)] = true
			}
			if inst.Src2 != 0 {
				escapedReps[find(inst.Src2)] = true
			}
		}
	}

	// 4. Determine which OpAlloc registers do not escape and are not already freed
	nonEscapingAllocRegs := make(map[uint32]uint16)
	for reg, typeID := range allocRegs {
		if !escapedReps[find(reg)] {
			nonEscapingAllocRegs[reg] = typeID
		}
	}

	if len(nonEscapingAllocRegs) == 0 {
		return false
	}

	// Remove registers that are already explicitly freed
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpFree && inst.Src1 != 0 {
			freedRep := find(inst.Src1)
			for reg := range nonEscapingAllocRegs {
				if find(reg) == freedRep {
					delete(nonEscapingAllocRegs, reg)
				}
			}
		}
	}

	if len(nonEscapingAllocRegs) == 0 {
		return false
	}

	// 5. Insert explicit OpFree for each non-escaping allocation right before each OpReturn instruction.
	changed := false

	// We iterate backwards to insert correctly without shifting unvisited indices.
	for i := len(fn.Insts) - 1; i >= 0; i-- {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpReturn {
			// Insert OpFree for all remaining non-escaping allocs
			var frees []air.AirInst
			for reg, typeID := range nonEscapingAllocRegs {
				frees = append(frees, air.AirInst{
					Opcode: air.OpFree,
					TypeID: typeID,
					Src1:   reg,
				})
			}

			if len(frees) == 0 {
				continue
			}

			// Slice and insert
			newInsts := make([]air.AirInst, 0, len(fn.Insts)+len(frees))
			newInsts = append(newInsts, fn.Insts[:i]...)
			newInsts = append(newInsts, frees...)
			newInsts = append(newInsts, fn.Insts[i:]...)
			fn.Insts = newInsts

			// Update all block instruction indices that point past this insertion point.
			updateBlockInstrs(fn, uint32(i-1), uint32(len(frees)))
			changed = true
		}
	}

	return changed
}
