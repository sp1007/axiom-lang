package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// CopyPropagationPass
//
// A forward dataflow optimization pass that propagates copy and move
// instructions. Under SSA form, since virtual registers are assigned exactly
// once, any copy `y = copy x` allows replacing all subsequent uses of `y`
// with `x` globally. Subsequent Dead Code Elimination (DCE) passes will then
// remove the redundant copy/move instructions.
// --------------------------------------------------------------------------

type CopyPropagationPass struct{}

func (p *CopyPropagationPass) Name() string { return "copy-prop" }

// Run executes the copy propagation pass on the given module.
func (p *CopyPropagationPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		if copyPropFunc(&mod.Funcs[fi]) {
			changed = true
		}
	}
	return changed
}

// copyPropFunc performs copy propagation on a single function.
func copyPropFunc(fn *air.AirFunc) bool {
	changed := false

	// Map: copy_vreg -> source_vreg
	copyMap := make(map[uint32]uint32)

	// Step 1: Collect all copies and moves
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}

		if (inst.Opcode == air.OpCopy || inst.Opcode == air.OpMove) && inst.Dest != 0 && inst.Src1 != 0 {
			// Save the copy binding
			copyMap[inst.Dest] = inst.Src1
		}
	}

	if len(copyMap) == 0 {
		return false
	}

	// Helper to resolve the root source register of a copy chain
	resolve := func(reg uint32) uint32 {
		if reg == 0 {
			return 0
		}
		curr := reg
		visited := make(map[uint32]bool) // detect cycles if any invalid IR is passed
		for {
			parent, exists := copyMap[curr]
			if !exists || visited[parent] {
				break
			}
			visited[curr] = true
			curr = parent
		}
		return curr
	}

	// Step 2: Propagate the copies to all instruction source operands
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}

		// Skip updating target inputs for instructions that define constants
		if inst.Opcode == air.OpIConst || inst.Opcode == air.OpFConst {
			continue
		}

		// Update Src1 if it is a register use (not a control target block ID)
		if inst.Src1 != 0 && inst.Opcode != air.OpJump {
			newSrc1 := resolve(inst.Src1)
			if newSrc1 != inst.Src1 {
				inst.Src1 = newSrc1
				changed = true
			}
		}

		// Update Src2 if it is a register use (not a control target block ID)
		if inst.Src2 != 0 && !inst.Opcode.IsControl() {
			newSrc2 := resolve(inst.Src2)
			if newSrc2 != inst.Src2 {
				inst.Src2 = newSrc2
				changed = true
			}
		}

		// Update Dest if used as a source (e.g. in OpStore or OpSetField)
		if inst.Dest != 0 && (inst.Opcode == air.OpStore || inst.Opcode == air.OpSetField) {
			newDest := resolve(inst.Dest)
			if newDest != inst.Dest {
				inst.Dest = newDest
				changed = true
			}
		}

		// Update condition register for branches
		if inst.Opcode == air.OpBranch && inst.Src1 != 0 {
			newCond := resolve(inst.Src1)
			if newCond != inst.Src1 {
				inst.Src1 = newCond
				changed = true
			}
		}

		// Update return value
		if inst.Opcode == air.OpReturn && inst.Src1 != 0 {
			newRet := resolve(inst.Src1)
			if newRet != inst.Src1 {
				inst.Src1 = newRet
				changed = true
			}
		}

		// Update arguments for function calls (stored in Extras)
		if inst.Opcode == air.OpCall {
			argStart := inst.Src2
			argCount := uint32(0)
			if argStart < uint32(len(fn.Extras)) {
				argCount = fn.Extras[argStart]
			}
			for idx := uint32(0); idx < argCount; idx++ {
				argRegIdx := argStart + 1 + idx
				argReg := fn.Extras[argRegIdx]
				if argReg != 0 {
					newArgReg := resolve(argReg)
					if newArgReg != argReg {
						fn.Extras[argRegIdx] = newArgReg
						changed = true
					}
				}
			}
		}
	}

	return changed
}
