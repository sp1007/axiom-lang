package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// Loop Unrolling Pass
//
// Unrolls small loops with a constant, small trip count (up to 4 iterations).
// This eliminates branch overhead, reduces control flow penalties, and
// enables other optimization passes (like GVN and Constant Folding) to optimize
// across iteration boundaries.
// --------------------------------------------------------------------------

type LoopUnrollPass struct{}

func (p *LoopUnrollPass) Name() string { return "loop-unroll" }

func (p *LoopUnrollPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		if unrollFunc(&mod.Funcs[fi]) {
			changed = true
		}
	}
	return changed
}

func unrollFunc(fn *air.AirFunc) bool {
	loops := detectLoops(fn)
	if len(loops) == 0 {
		return false
	}

	changed := false
	// Process loops in reverse order (innermost first) to handle nested loops correctly
	for i := len(loops) - 1; i >= 0; i-- {
		if tryUnrollLoop(fn, &loops[i]) {
			changed = true
		}
	}
	return changed
}

// tryUnrollLoop attempts to unroll a detected natural loop.
// Returns true if the loop was successfully unrolled.
func tryUnrollLoop(fn *air.AirFunc, loop *LoopInfo) bool {
	var isSingleBlock bool
	var headerBlk *air.BasicBlock
	var bodyBlk *air.BasicBlock

	if loop.HeaderBlock == loop.BackEdge {
		isSingleBlock = true
		headerBlk = &fn.Blocks[loop.HeaderBlock]
		bodyBlk = headerBlk
	} else if len(loop.BodyBlocks) == 2 {
		isSingleBlock = false
		headerBlk = &fn.Blocks[loop.HeaderBlock]
		bodyBlk = &fn.Blocks[loop.BackEdge]
	} else {
		// More complex loops are not supported for unrolling
		return false
	}

	// Loop body must be small
	if len(bodyBlk.Instrs) > 15 {
		return false
	}

	// Identify the loop induction limit.
	// Typically, a loop has a comparison in the headerBlock: %cond = le/lt %i, LIMIT
	// and a branch: branch %cond, body_block, exit_block
	var condReg uint32
	var exitBlock uint32
	var hasBranch bool

	for _, instIdx := range headerBlk.Instrs {
		if int(instIdx) < len(fn.Insts) {
			inst := &fn.Insts[instIdx]
			if inst.Opcode == air.OpBranch {
				condReg = inst.Src1
				exitBlock = inst.Dest // false target
				hasBranch = true
				break
			}
		}
	}

	if !hasBranch || condReg == 0 || len(loop.ExitBlocks) == 0 {
		return false
	}

	// Verify that the condition is compared against a small constant limit.
	var limit uint32
	var hasLimit bool

	for _, instIdx := range headerBlk.Instrs {
		if int(instIdx) < len(fn.Insts) {
			inst := &fn.Insts[instIdx]
			if inst.Dest == condReg && inst.Opcode.IsBinaryALU() {
				// Check if either operand is an integer constant
				if isConstReg(fn, inst.Src1) {
					limit = getConstVal(fn, inst.Src1)
					hasLimit = true
				} else if isConstReg(fn, inst.Src2) {
					limit = getConstVal(fn, inst.Src2)
					hasLimit = true
				}
			}
		}
	}

	// Only unroll if the trip count is small (limit <= 4)
	if !hasLimit || limit == 0 || limit > 4 {
		return false
	}

	numIterations := int(limit)

	if isSingleBlock {
		var newBlocks []air.BasicBlock

		// First copy of blocks will replace the original header block
		// We will create new basic blocks for iterations 2..K.
		firstBlockID := loop.HeaderBlock

		// Keep track of the predecessor of the next unrolled block.
		prevBlockID := firstBlockID

		// Track the next register ID generator
		maxReg := uint32(0)
		for i := range fn.Insts {
			if fn.Insts[i].Dest > maxReg {
				maxReg = fn.Insts[i].Dest
			}
		}
		freshReg := func() uint32 {
			maxReg++
			return maxReg
		}

		// Retrieve original instructions of the loop body
		origInstrs := make([]air.AirInst, 0, len(headerBlk.Instrs))
		for _, idx := range headerBlk.Instrs {
			if int(idx) < len(fn.Insts) {
				origInstrs = append(origInstrs, fn.Insts[idx])
			}
		}

		// Create unrolled blocks
		for iter := 1; iter < numIterations; iter++ {
			nextBlockID := uint32(len(fn.Blocks)) + uint32(len(newBlocks))

			// Map registers from original body to fresh registers for this iteration
			regMap := make(map[uint32]uint32)

			var newInstrs []uint32
			for _, inst := range origInstrs {
				if inst.Opcode == air.OpNop {
					continue
				}

				// Don't copy the terminator branch — we replace it with a direct jump to the next unrolled block!
				if inst.Opcode == air.OpBranch || inst.Opcode == air.OpJump {
					continue
				}

				// Create fresh registers for destinations (if strictly defined once)
				if inst.Dest != 0 && inst.Opcode != air.OpCopy {
					newReg := freshReg()
					regMap[inst.Dest] = newReg
					inst.Dest = newReg
				}

				// Remap source operands
				if r, mapped := regMap[inst.Src1]; mapped {
					inst.Src1 = r
				}
				if r, mapped := regMap[inst.Src2]; mapped {
					inst.Src2 = r
				}

				instIdx := uint32(len(fn.Insts))
				fn.Insts = append(fn.Insts, inst)
				newInstrs = append(newInstrs, instIdx)
			}

			// Emit the direct jump to the next iteration block
			jumpIdx := uint32(len(fn.Insts))
			fn.Insts = append(fn.Insts, air.AirInst{
				Opcode: air.OpJump,
				Src1:   nextBlockID + 1, // point to next iteration block (or exit)
			})
			newInstrs = append(newInstrs, jumpIdx)

			newBlk := air.BasicBlock{
				ID:     nextBlockID,
				Instrs: newInstrs,
				Succs:  []uint32{nextBlockID + 1},
				Preds:  []uint32{prevBlockID},
			}
			newBlocks = append(newBlocks, newBlk)
			prevBlockID = nextBlockID
		}

		// Adjust the original loop header block (first iteration):
		// - Remove the branch instruction and replace it with a direct jump to the second iteration.
		var adjustedInstrs []uint32
		for _, idx := range headerBlk.Instrs {
			if int(idx) < len(fn.Insts) {
				inst := &fn.Insts[idx]
				if inst.Opcode != air.OpBranch && inst.Opcode != air.OpJump {
					adjustedInstrs = append(adjustedInstrs, idx)
				}
			}
		}

		nextIterBlockID := firstBlockID + 1
		if numIterations == 1 {
			nextIterBlockID = exitBlock
		}

		jumpIdx := uint32(len(fn.Insts))
		fn.Insts = append(fn.Insts, air.AirInst{
			Opcode: air.OpJump,
			Src1:   nextIterBlockID,
		})
		adjustedInstrs = append(adjustedInstrs, jumpIdx)
		headerBlk.Instrs = adjustedInstrs
		headerBlk.Succs = []uint32{nextIterBlockID}

		// Add the unrolled blocks to the function
		for _, blk := range newBlocks {
			// Adjust the last jump's target if it's the last iteration block
			if blk.ID == prevBlockID {
				lastInstIdx := blk.Instrs[len(blk.Instrs)-1]
				fn.Insts[lastInstIdx].Src1 = exitBlock
				blk.Succs = []uint32{exitBlock}
			}
			fn.Blocks = append(fn.Blocks, blk)
		}

		// Update predecessor/successor links for exit block
		if int(exitBlock) < len(fn.Blocks) {
			exitBlk := &fn.Blocks[exitBlock]
			// Replace original header reference in Preds list with the last unrolled block
			var newPreds []uint32
			for _, pred := range exitBlk.Preds {
				if pred != loop.HeaderBlock {
					newPreds = append(newPreds, pred)
				}
			}
			newPreds = append(newPreds, prevBlockID)
			exitBlk.Preds = newPreds
		}
	} else {
		// Multi-block loop (isSingleBlock == false)
		var newBlocks []air.BasicBlock
		firstBodyBlockID := loop.BackEdge
		prevBlockID := firstBodyBlockID

		// Track the next register ID generator
		maxReg := uint32(0)
		for i := range fn.Insts {
			if fn.Insts[i].Dest > maxReg {
				maxReg = fn.Insts[i].Dest
			}
		}
		freshReg := func() uint32 {
			maxReg++
			return maxReg
		}

		// Retrieve original instructions of the loop body block (bodyBlk)
		origInstrs := make([]air.AirInst, 0, len(bodyBlk.Instrs))
		for _, idx := range bodyBlk.Instrs {
			if int(idx) < len(fn.Insts) {
				origInstrs = append(origInstrs, fn.Insts[idx])
			}
		}

		// Create unrolled blocks for iterations 2..K.
		for iter := 1; iter < numIterations; iter++ {
			nextBlockID := uint32(len(fn.Blocks)) + uint32(len(newBlocks))

			// Map registers from original body to fresh registers for this iteration
			regMap := make(map[uint32]uint32)

			var newInstrs []uint32
			for _, inst := range origInstrs {
				if inst.Opcode == air.OpNop {
					continue
				}

				// Don't copy the terminator jump back to header — we replace it with a direct jump to the next block!
				if inst.Opcode == air.OpJump || inst.Opcode == air.OpBranch {
					continue
				}

				// Create fresh registers for destinations (if strictly defined once)
				if inst.Dest != 0 && inst.Opcode != air.OpCopy {
					newReg := freshReg()
					regMap[inst.Dest] = newReg
					inst.Dest = newReg
				}

				// Remap source operands
				if r, mapped := regMap[inst.Src1]; mapped {
					inst.Src1 = r
				}
				if r, mapped := regMap[inst.Src2]; mapped {
					inst.Src2 = r
				}

				instIdx := uint32(len(fn.Insts))
				fn.Insts = append(fn.Insts, inst)
				newInstrs = append(newInstrs, instIdx)
			}

			// Emit the direct jump to the next iteration block
			jumpIdx := uint32(len(fn.Insts))
			fn.Insts = append(fn.Insts, air.AirInst{
				Opcode: air.OpJump,
				Src1:   nextBlockID + 1, // point to next iteration block (or exit)
			})
			newInstrs = append(newInstrs, jumpIdx)

			newBlk := air.BasicBlock{
				ID:     nextBlockID,
				Instrs: newInstrs,
				Succs:  []uint32{nextBlockID + 1},
				Preds:  []uint32{prevBlockID},
			}
			newBlocks = append(newBlocks, newBlk)
			prevBlockID = nextBlockID
		}

		// Adjust the original loop header block (condition block):
		// - Remove the branch instruction and replace it with a direct jump to the first body block.
		var adjustedHeaderInstrs []uint32
		for _, idx := range headerBlk.Instrs {
			if int(idx) < len(fn.Insts) {
				inst := &fn.Insts[idx]
				if inst.Opcode != air.OpBranch && inst.Opcode != air.OpJump {
					adjustedHeaderInstrs = append(adjustedHeaderInstrs, idx)
				}
			}
		}

		headerJumpIdx := uint32(len(fn.Insts))
		fn.Insts = append(fn.Insts, air.AirInst{
			Opcode: air.OpJump,
			Src1:   firstBodyBlockID,
		})
		adjustedHeaderInstrs = append(adjustedHeaderInstrs, headerJumpIdx)
		headerBlk.Instrs = adjustedHeaderInstrs
		headerBlk.Succs = []uint32{firstBodyBlockID}

		// Remove loop.BackEdge from headerBlk's Preds (since the backedge is removed)
		var newHeaderPreds []uint32
		for _, pred := range headerBlk.Preds {
			if pred != loop.BackEdge {
				newHeaderPreds = append(newHeaderPreds, pred)
			}
		}
		headerBlk.Preds = newHeaderPreds

		// Adjust the first body block (first iteration):
		// - Remove the jump back to the header and replace it with a jump to the second iteration (or exit if limit is 1).
		var adjustedBodyInstrs []uint32
		for _, idx := range bodyBlk.Instrs {
			if int(idx) < len(fn.Insts) {
				inst := &fn.Insts[idx]
				if inst.Opcode != air.OpJump && inst.Opcode != air.OpBranch {
					adjustedBodyInstrs = append(adjustedBodyInstrs, idx)
				}
			}
		}

		nextIterBlockID := uint32(0)
		if numIterations > 1 {
			nextIterBlockID = uint32(len(fn.Blocks))
		} else {
			nextIterBlockID = exitBlock
		}

		bodyJumpIdx := uint32(len(fn.Insts))
		fn.Insts = append(fn.Insts, air.AirInst{
			Opcode: air.OpJump,
			Src1:   nextIterBlockID,
		})
		adjustedBodyInstrs = append(adjustedBodyInstrs, bodyJumpIdx)
		bodyBlk.Instrs = adjustedBodyInstrs
		bodyBlk.Succs = []uint32{nextIterBlockID}
		bodyBlk.Preds = []uint32{loop.HeaderBlock}

		// Add the unrolled blocks to the function
		for _, blk := range newBlocks {
			// Adjust the last jump's target if it's the last iteration block
			if blk.ID == prevBlockID {
				lastInstIdx := blk.Instrs[len(blk.Instrs)-1]
				fn.Insts[lastInstIdx].Src1 = exitBlock
				blk.Succs = []uint32{exitBlock}
			}
			fn.Blocks = append(fn.Blocks, blk)
		}

		// Update predecessor/successor links for exit block
		if int(exitBlock) < len(fn.Blocks) {
			exitBlk := &fn.Blocks[exitBlock]
			var newPreds []uint32
			for _, pred := range exitBlk.Preds {
				if pred != loop.HeaderBlock && pred != loop.BackEdge {
					newPreds = append(newPreds, pred)
				}
			}
			newPreds = append(newPreds, prevBlockID)
			exitBlk.Preds = newPreds
		}
	}

	return true
}

func isConstReg(fn *air.AirFunc, reg uint32) bool {
	if reg == 0 {
		return false
	}
	for i := range fn.Insts {
		if fn.Insts[i].Dest == reg && (fn.Insts[i].Opcode == air.OpIConst || fn.Insts[i].Opcode == air.OpFConst) {
			return true
		}
	}
	return false
}

func getConstVal(fn *air.AirFunc, reg uint32) uint32 {
	for i := range fn.Insts {
		if fn.Insts[i].Dest == reg && (fn.Insts[i].Opcode == air.OpIConst || fn.Insts[i].Opcode == air.OpFConst) {
			return fn.Insts[i].Src1
		}
	}
	return 0
}
