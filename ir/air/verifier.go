// Package air — AIR verification pass.
//
// The verifier checks structural invariants of an AirFunc and reports
// ALL violations (it never fails fast). This is intended to be run
// after every IR transformation to catch bugs early.
package air

import "fmt"

// ---------------------------------------------------------------------------
// VerifyError — a single verification failure.
// ---------------------------------------------------------------------------

// VerifyError represents a single verification failure pinpointed to a
// specific block and instruction index within that block.
type VerifyError struct {
	BlockID uint32 // block where the error was found
	InstIdx uint32 // instruction index within the block (0-based), ^0 for block-level
	Message string // human-readable description
}

// Error implements the error interface.
func (e *VerifyError) Error() string {
	if e.InstIdx == ^uint32(0) {
		return fmt.Sprintf("block_%d: %s", e.BlockID, e.Message)
	}
	return fmt.Sprintf("block_%d[%d]: %s", e.BlockID, e.InstIdx, e.Message)
}

// ---------------------------------------------------------------------------
// Verify — check all invariants of an AirFunc.
// ---------------------------------------------------------------------------

// Verify checks all structural invariants of an AirFunc and returns every
// error found. An empty slice means the function is well-formed.
//
// Checks performed:
//  1. SSA: each Dest register defined at most once across all instructions
//  2. Terminators: every non-empty block ends with exactly one terminator;
//     no terminators appear before the last instruction
//  3. Entry block: block 0 has IsEntry=true and no predecessors
//  4. Phi placement: phi nodes appear only at the start of a block
//  5. Branch targets: OpJump and OpBranch reference valid block IDs
//  6. Successor consistency: terminator targets match the block's Succs list
//  7. Empty blocks: warning (not error) if a block has no instructions
func Verify(fn *AirFunc) []VerifyError {
	var errs []VerifyError

	add := func(bid, iidx uint32, msg string) {
		errs = append(errs, VerifyError{BlockID: bid, InstIdx: iidx, Message: msg})
	}

	numBlocks := uint32(len(fn.Blocks))

	// -----------------------------------------------------------------------
	// Check 3: Entry block
	// -----------------------------------------------------------------------
	if numBlocks > 0 {
		entry := &fn.Blocks[0]
		if !entry.IsEntry {
			add(0, ^uint32(0), "block 0 must have IsEntry=true")
		}
		if len(entry.Preds) > 0 {
			add(0, ^uint32(0), fmt.Sprintf(
				"entry block must have no predecessors, has %d", len(entry.Preds)))
		}
	}

	// SSA def tracking: register → (blockID, instIdx within block)
	defs := make(map[uint32]struct{})

	for bi := uint32(0); bi < numBlocks; bi++ {
		blk := &fn.Blocks[bi]

		// -------------------------------------------------------------------
		// Check 7: Empty blocks (warning)
		// -------------------------------------------------------------------
		if len(blk.Instrs) == 0 {
			add(bi, ^uint32(0), "block has no instructions (warning)")
			continue
		}

		phiDone := false // tracks whether we've seen a non-phi instruction

		for ii, instIdx := range blk.Instrs {
			if int(instIdx) >= len(fn.Insts) {
				add(bi, uint32(ii), fmt.Sprintf(
					"instruction index %d out of range (len=%d)", instIdx, len(fn.Insts)))
				continue
			}
			inst := fn.Insts[instIdx]

			// Skip NOPs entirely — they don't participate in SSA or control flow.
			if inst.Opcode == OpNop {
				continue
			}

			// ---------------------------------------------------------------
			// Check 1: SSA — each Dest defined at most once
			// ---------------------------------------------------------------
			if inst.Dest != 0 && !isDestUsedAsOperand(inst.Opcode) {
				if _, dup := defs[inst.Dest]; dup {
					add(bi, uint32(ii), fmt.Sprintf(
						"SSA violation: register %%%d defined more than once", inst.Dest))
				} else {
					defs[inst.Dest] = struct{}{}
				}
			}

			// ---------------------------------------------------------------
			// Check 4: Phi placement — phis must precede all non-phis
			// ---------------------------------------------------------------
			if inst.Opcode == OpPhi {
				if phiDone {
					add(bi, uint32(ii), "phi instruction after non-phi instruction")
				}
			} else {
				phiDone = true
			}

			// ---------------------------------------------------------------
			// Check 2: Terminators — only at the end of the block
			// ---------------------------------------------------------------
			isLast := ii == len(blk.Instrs)-1
			if inst.Opcode.IsTerminator() {
				if !isLast {
					add(bi, uint32(ii), fmt.Sprintf(
						"terminator %s is not the last instruction in the block",
						inst.Opcode.Mnemonic()))
				}
			} else if isLast {
				add(bi, uint32(ii), fmt.Sprintf(
					"last instruction %s is not a terminator",
					inst.Opcode.Mnemonic()))
			}

			// ---------------------------------------------------------------
			// Checks 5 & 6: Branch targets and successor consistency
			// ---------------------------------------------------------------
			if inst.Opcode == OpJump {
				target := inst.Src1
				if target >= numBlocks {
					add(bi, uint32(ii), fmt.Sprintf(
						"jump target block_%d is out of range (num_blocks=%d)",
						target, numBlocks))
				}
				if isLast {
					checkSuccConsistency(&errs, blk, bi, uint32(ii), []uint32{target})
				}
			}
			if inst.Opcode == OpBranch {
				trueTarget := inst.Src2
				falseTarget := inst.Dest
				if trueTarget >= numBlocks {
					add(bi, uint32(ii), fmt.Sprintf(
						"branch true target block_%d is out of range (num_blocks=%d)",
						trueTarget, numBlocks))
				}
				if falseTarget >= numBlocks {
					add(bi, uint32(ii), fmt.Sprintf(
						"branch false target block_%d is out of range (num_blocks=%d)",
						falseTarget, numBlocks))
				}
				if isLast {
					targets := []uint32{trueTarget, falseTarget}
					// deduplicate if both targets are the same block
					if trueTarget == falseTarget {
						targets = []uint32{trueTarget}
					}
					checkSuccConsistency(&errs, blk, bi, uint32(ii), targets)
				}
			}
		}
	}

	return errs
}

// checkSuccConsistency verifies that the terminator's targets exactly match
// the block's Succs list (order-independent).
func checkSuccConsistency(errs *[]VerifyError, blk *BasicBlock, bid, iidx uint32, targets []uint32) {
	succs := blk.Succs

	// Build sets for comparison.
	targetSet := make(map[uint32]bool, len(targets))
	for _, t := range targets {
		targetSet[t] = true
	}
	succSet := make(map[uint32]bool, len(succs))
	for _, s := range succs {
		succSet[s] = true
	}

	// Check targets ⊆ succs
	for _, t := range targets {
		if !succSet[t] {
			*errs = append(*errs, VerifyError{
				BlockID: bid,
				InstIdx: iidx,
				Message: fmt.Sprintf(
					"terminator targets block_%d but it is not in Succs %v", t, succs),
			})
		}
	}
	// Check succs ⊆ targets
	for _, s := range succs {
		if !targetSet[s] {
			*errs = append(*errs, VerifyError{
				BlockID: bid,
				InstIdx: iidx,
				Message: fmt.Sprintf(
					"Succs contains block_%d but terminator does not target it", s),
			})
		}
	}
}

// isDestUsedAsOperand returns true for opcodes where the Dest field is used
// as an operand rather than as an SSA definition. OpBranch uses Dest for
// the false target block ID.
func isDestUsedAsOperand(op Opcode) bool {
	return op == OpBranch
}
