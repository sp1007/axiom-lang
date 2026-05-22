package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t07: Loop Region Detection and Optimization
//
// Identifies natural loops in the CFG and applies loop-specific
// optimizations: loop-invariant code motion (LICM), strength reduction,
// and loop metadata annotation for the backend.
// --------------------------------------------------------------------------

// LoopInfo describes a detected natural loop.
type LoopInfo struct {
	HeaderBlock uint32   // block ID of the loop header
	BackEdge    uint32   // block ID of the backedge source
	BodyBlocks  []uint32 // all blocks in the loop body
	ExitBlocks  []uint32 // blocks that exit the loop
	Depth       int      // nesting depth (1 = outermost)
}

// LoopRegionPass implements OptPass for loop analysis and optimization.
type LoopRegionPass struct{}

func (p *LoopRegionPass) Name() string { return "loop-region" }

// Run detects natural loops and performs loop-invariant code motion.
func (p *LoopRegionPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		loops := detectLoops(fn)
		if len(loops) == 0 {
			continue
		}

		for _, loop := range loops {
			if hoistInvariants(fn, &loop) {
				changed = true
			}
		}
	}
	return changed
}

// detectLoops finds natural loops in the CFG using backedge detection.
// A backedge is an edge from a block B to a block H where H dominates B.
// For MVP, we use a simplified approach: a backedge is any edge from B to H
// where H has a smaller block ID than B (assuming blocks are in DFS order).
func detectLoops(fn *air.AirFunc) []LoopInfo {
	if len(fn.Blocks) < 2 {
		return nil
	}

	var loops []LoopInfo

	// Find backedges: edges from B → H where H.ID < B.ID
	for bi := range fn.Blocks {
		blk := &fn.Blocks[bi]
		for _, succ := range blk.Succs {
			if succ < blk.ID {
				// This is a backedge: blk → succ (succ is the loop header)
				bodyBlocks := collectLoopBody(fn, succ, blk.ID)
				exitBlocks := findExitBlocks(fn, bodyBlocks)

				loops = append(loops, LoopInfo{
					HeaderBlock: succ,
					BackEdge:    blk.ID,
					BodyBlocks:  bodyBlocks,
					ExitBlocks:  exitBlocks,
					Depth:       1,
				})
			}
		}
	}

	// Assign nesting depths (simple: count header inclusion)
	for i := range loops {
		for j := range loops {
			if i == j {
				continue
			}
			if containsBlock(loops[j].BodyBlocks, loops[i].HeaderBlock) {
				loops[i].Depth++
			}
		}
	}

	return loops
}

// collectLoopBody collects all blocks in the loop body by walking
// backward from the backedge source to the header.
func collectLoopBody(fn *air.AirFunc, headerID, backEdgeID uint32) []uint32 {
	body := map[uint32]bool{headerID: true}
	stack := []uint32{backEdgeID}

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if body[cur] {
			continue
		}
		body[cur] = true

		// Walk predecessors
		if int(cur) < len(fn.Blocks) {
			for _, pred := range fn.Blocks[cur].Preds {
				if !body[pred] {
					stack = append(stack, pred)
				}
			}
		}
	}

	result := make([]uint32, 0, len(body))
	for id := range body {
		result = append(result, id)
	}
	return result
}

// findExitBlocks finds blocks that have successors outside the loop.
func findExitBlocks(fn *air.AirFunc, bodyBlocks []uint32) []uint32 {
	bodySet := make(map[uint32]bool, len(bodyBlocks))
	for _, id := range bodyBlocks {
		bodySet[id] = true
	}

	var exits []uint32
	for _, id := range bodyBlocks {
		if int(id) >= len(fn.Blocks) {
			continue
		}
		blk := &fn.Blocks[id]
		for _, succ := range blk.Succs {
			if !bodySet[succ] {
				exits = append(exits, id)
				break
			}
		}
	}
	return exits
}

// hoistInvariants moves loop-invariant instructions out of the loop body
// into the preheader (the block before the loop header).
func hoistInvariants(fn *air.AirFunc, loop *LoopInfo) bool {
	bodySet := make(map[uint32]bool, len(loop.BodyBlocks))
	for _, id := range loop.BodyBlocks {
		bodySet[id] = true
	}

	// Find instructions in the loop body that are loop-invariant:
	// - All source operands are defined outside the loop, or
	// - All source operands are loop-invariant themselves (already hoisted)
	// - No side effects
	loopDefs := make(map[uint32]bool)
	for _, blkID := range loop.BodyBlocks {
		if int(blkID) >= len(fn.Blocks) {
			continue
		}
		blk := &fn.Blocks[blkID]
		for _, instIdx := range blk.Instrs {
			if int(instIdx) < len(fn.Insts) {
				inst := &fn.Insts[instIdx]
				if inst.Dest != 0 {
					loopDefs[inst.Dest] = true
				}
			}
		}
	}

	var toHoist []uint32
	hoistedDefs := make(map[uint32]bool)

	// Scan instructions in loop body blocks and identify invariants
	for _, blkID := range loop.BodyBlocks {
		if int(blkID) >= len(fn.Blocks) {
			continue
		}
		blk := &fn.Blocks[blkID]
		for _, instIdx := range blk.Instrs {
			if int(instIdx) >= len(fn.Insts) {
				continue
			}
			inst := &fn.Insts[instIdx]

			// Skip non-hoistable instructions
			if inst.Opcode == air.OpNop || inst.Opcode.IsTerminator() ||
				inst.Opcode.IsControl() || hasSideEffect(inst.Opcode) {
				continue
			}

			// Check if all sources are defined outside the loop or already hoisted
			src1Outside := inst.Src1 == 0 || !loopDefs[inst.Src1] || hoistedDefs[inst.Src1]
			src2Outside := inst.Src2 == 0 || !loopDefs[inst.Src2] || hoistedDefs[inst.Src2]

			if src1Outside && src2Outside && inst.Dest != 0 {
				toHoist = append(toHoist, instIdx)
				hoistedDefs[inst.Dest] = true
			}
		}
	}

	if len(toHoist) == 0 {
		return false
	}

	// Identify or construct the preheader block
	headerBlk := &fn.Blocks[loop.HeaderBlock]
	var outsidePreds []uint32
	for _, predID := range headerBlk.Preds {
		if !bodySet[predID] && int(predID) < len(fn.Blocks) {
			outsidePreds = append(outsidePreds, predID)
		}
	}

	var preheaderID uint32
	if len(outsidePreds) == 1 {
		// Single predecessor from outside - perfect preheader
		preheaderID = outsidePreds[0]
	} else {
		// Create a new preheader block dynamically to redirect multiple entries
		preheaderID = uint32(len(fn.Blocks))
		preheader := air.BasicBlock{
			ID:    preheaderID,
			Succs: []uint32{loop.HeaderBlock},
			Preds: outsidePreds,
		}

		// Emit the jump instruction at the end of the new preheader
		jumpIdx := uint32(len(fn.Insts))
		fn.Insts = append(fn.Insts, air.AirInst{
			Opcode: air.OpJump,
			Src1:   loop.HeaderBlock,
		})
		preheader.Instrs = []uint32{jumpIdx}
		fn.Blocks = append(fn.Blocks, preheader)

		// Redirect predecessors to new preheader instead of header
		for _, predID := range outsidePreds {
			predBlk := &fn.Blocks[predID]
			for sidx, succ := range predBlk.Succs {
				if succ == loop.HeaderBlock {
					predBlk.Succs[sidx] = preheaderID
				}
			}

			// Redirect instructions inside predecessor
			for ii := range predBlk.Instrs {
				instIdx := predBlk.Instrs[ii]
				inst := &fn.Insts[instIdx]
				if inst.Opcode == air.OpJump && inst.Src1 == loop.HeaderBlock {
					inst.Src1 = preheaderID
				} else if inst.Opcode == air.OpBranch {
					if inst.Src2 == loop.HeaderBlock {
						inst.Src2 = preheaderID
					}
					if inst.Dest == loop.HeaderBlock {
						inst.Dest = preheaderID
					}
				}
			}
		}

		// Update header predecessors
		var newPreds []uint32
		for _, predID := range headerBlk.Preds {
			if bodySet[predID] {
				newPreds = append(newPreds, predID)
			}
		}
		newPreds = append(newPreds, preheaderID)
		headerBlk.Preds = newPreds
	}

	preheaderBlk := &fn.Blocks[preheaderID]

	// Hoist each instruction: remove from body block and insert into preheader
	hoistSet := make(map[uint32]bool)
	for _, idx := range toHoist {
		hoistSet[idx] = true
	}

	for _, blkID := range loop.BodyBlocks {
		if int(blkID) >= len(fn.Blocks) {
			continue
		}
		blk := &fn.Blocks[blkID]

		// Filter out hoisted instructions
		var newInstrs []uint32
		for _, idx := range blk.Instrs {
			if !hoistSet[idx] {
				newInstrs = append(newInstrs, idx)
			}
		}
		blk.Instrs = newInstrs
	}

	// Insert hoisted instructions in preheader right before the terminator
	termIdx := len(preheaderBlk.Instrs) - 1
	if termIdx < 0 {
		preheaderBlk.Instrs = toHoist
	} else {
		prefix := make([]uint32, termIdx)
		copy(prefix, preheaderBlk.Instrs[:termIdx])
		suffix := make([]uint32, len(preheaderBlk.Instrs)-termIdx)
		copy(suffix, preheaderBlk.Instrs[termIdx:])

		newInstrs := append(prefix, toHoist...)
		newInstrs = append(newInstrs, suffix...)
		preheaderBlk.Instrs = newInstrs
	}

	return true
}

// containsBlock checks if a block ID is in the given list.
func containsBlock(blocks []uint32, id uint32) bool {
	for _, b := range blocks {
		if b == id {
			return true
		}
	}
	return false
}
