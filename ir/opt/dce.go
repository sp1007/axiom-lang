package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t03: Dead Code Elimination Pass
//
// Removes instructions whose results are never used and have no side effects.
// Also removes unreachable blocks (blocks not reachable from the entry block).
// --------------------------------------------------------------------------

// DCEPass implements OptPass for dead code elimination.
type DCEPass struct{}

func (p *DCEPass) Name() string { return "dce" }

// Run performs dead code elimination on all functions in the module.
// Returns true if any instruction was removed (replaced with OpNop).
func (p *DCEPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		if dceFunc(&mod.Funcs[fi]) {
			changed = true
		}
	}
	return changed
}

// dceFunc performs DCE on a single function.
func dceFunc(fn *air.AirFunc) bool {
	changed := false

	// Step 1: Count uses of each register
	useCount := buildUseCount(fn)

	// Step 2: Mark instructions with zero uses and no side effects as NOP
	for i := range fn.Insts {
		inst := &fn.Insts[i]

		// Skip already-NOPed instructions
		if inst.Opcode == air.OpNop {
			continue
		}

		// Skip instructions that don't define a value
		if inst.Dest == 0 {
			continue
		}

		// Skip instructions with side effects — they must stay
		if hasSideEffect(inst.Opcode) {
			continue
		}

		// If no one uses the result, eliminate the instruction
		if useCount[inst.Dest] == 0 {
			inst.Opcode = air.OpNop
			inst.TypeID = 0
			inst.Dest = 0
			inst.Src1 = 0
			inst.Src2 = 0
			changed = true
		}
	}

	// Step 3: Remove unreachable blocks
	if removeUnreachableBlocks(fn) {
		changed = true
	}

	return changed
}

// buildUseCount counts how many times each register is used as a source operand.
func buildUseCount(fn *air.AirFunc) map[uint32]int {
	uses := make(map[uint32]int, len(fn.Insts))

	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}

		// Count Src1 usage
		if inst.Src1 != 0 {
			uses[inst.Src1]++
		}

		// Count Src2 usage (skip for branch/jump where Src2 is a block target)
		if inst.Src2 != 0 {
			if inst.Opcode != air.OpBranch && inst.Opcode != air.OpJump {
				uses[inst.Src2]++
			}
		}

		// For branch: Src1 (condition) is a register use, Src2 and Dest are block targets
		if inst.Opcode == air.OpBranch {
			// Src1 is the condition register (already counted above)
			// Src2 and Dest are block targets, not register uses
		}

		// For return: Src1 is the return value register
		if inst.Opcode == air.OpReturn {
			// Already counted above
		}
	}

	return uses
}

// hasSideEffect returns true if the opcode has observable side effects
// and should never be eliminated, even if its result is unused.
func hasSideEffect(op air.Opcode) bool {
	switch op {
	case air.OpStore, air.OpFree, air.OpDestroy, air.OpAliasReuse:
		return true
	case air.OpCall, air.OpSpawn, air.OpSend:
		return true
	case air.OpReturn, air.OpJump, air.OpBranch:
		return true
	case air.OpSetField:
		return true
	default:
		return false
	}
}

// removeUnreachableBlocks removes blocks not reachable from the entry block.
// Returns true if any blocks were removed.
func removeUnreachableBlocks(fn *air.AirFunc) bool {
	if len(fn.Blocks) <= 1 {
		return false
	}

	// BFS from entry block (block 0)
	reachable := make(map[uint32]bool, len(fn.Blocks))
	queue := []uint32{0}
	reachable[0] = true

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if int(cur) >= len(fn.Blocks) {
			continue
		}
		blk := &fn.Blocks[cur]
		for _, succ := range blk.Succs {
			if !reachable[succ] {
				reachable[succ] = true
				queue = append(queue, succ)
			}
		}
	}

	// Remove unreachable blocks
	if len(reachable) == len(fn.Blocks) {
		return false // all blocks are reachable
	}

	// NOP out instructions in unreachable blocks
	for bi := range fn.Blocks {
		blk := &fn.Blocks[bi]
		if !reachable[blk.ID] {
			for _, instIdx := range blk.Instrs {
				if int(instIdx) < len(fn.Insts) {
					fn.Insts[instIdx] = air.AirInst{Opcode: air.OpNop}
				}
			}
			blk.Instrs = nil
			blk.Succs = nil
			blk.Preds = nil
		}
	}

	return true
}
