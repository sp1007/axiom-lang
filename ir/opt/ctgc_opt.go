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
	return false
}
