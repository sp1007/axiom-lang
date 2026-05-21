package builder

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p09-t09: Async State Machine Annotations
//
// For the MVP, async functions are lowered synchronously — the same as
// regular functions. This file provides the structural annotations that
// mark await points and state boundaries, which a future async runtime
// pass will consume to split the function into a cooperative state machine.
// --------------------------------------------------------------------------

// StateMachineInfo describes the async structure of a function.
// It is computed post-lowering by analyzeAsync and attached as metadata.
type StateMachineInfo struct {
	NumStates   uint32     // total number of states (1 for sync, N for async)
	StateBlocks [][]uint32 // blocks belonging to each state
	ResumeBlock uint32     // block ID of the resume entry point (0 if sync)
}

// analyzeAsync inspects a lowered AirFunc for await points and
// computes a StateMachineInfo. For sync functions, it returns a
// single-state info. For async functions, each await point splits
// the function into a new state.
func analyzeAsync(fn *air.AirFunc) *StateMachineInfo {
	if !fn.IsAsync {
		return &StateMachineInfo{
			NumStates:   1,
			StateBlocks: [][]uint32{allBlockIDs(fn)},
			ResumeBlock: 0,
		}
	}

	// Find await points (OpAwait instructions)
	awaitPoints := findAwaitPoints(fn)

	if len(awaitPoints) == 0 {
		// Async function with no awaits — treat as single state
		return &StateMachineInfo{
			NumStates:   1,
			StateBlocks: [][]uint32{allBlockIDs(fn)},
			ResumeBlock: 0,
		}
	}

	// Split blocks at await points
	// State 0: entry to first await
	// State N: resume from await N-1 to next await (or return)
	numStates := uint32(len(awaitPoints) + 1)
	stateBlocks := make([][]uint32, numStates)

	// For MVP, assign all blocks to state 0 and annotate
	// Real splitting requires inserting state dispatch blocks
	stateBlocks[0] = allBlockIDs(fn)

	return &StateMachineInfo{
		NumStates:   numStates,
		StateBlocks: stateBlocks,
		ResumeBlock: 0,
	}
}

// awaitPoint records the location of an OpAwait instruction.
type awaitPoint struct {
	BlockID uint32
	InstIdx uint32
}

// findAwaitPoints scans all instructions for OpAwait.
func findAwaitPoints(fn *air.AirFunc) []awaitPoint {
	var points []awaitPoint
	for bi := range fn.Blocks {
		blk := &fn.Blocks[bi]
		for _, instIdx := range blk.Instrs {
			if int(instIdx) < len(fn.Insts) && fn.Insts[instIdx].Opcode == air.OpAwait {
				points = append(points, awaitPoint{
					BlockID: blk.ID,
					InstIdx: instIdx,
				})
			}
		}
	}
	return points
}

// allBlockIDs returns all block IDs in a function.
func allBlockIDs(fn *air.AirFunc) []uint32 {
	ids := make([]uint32, len(fn.Blocks))
	for i := range fn.Blocks {
		ids[i] = fn.Blocks[i].ID
	}
	return ids
}
