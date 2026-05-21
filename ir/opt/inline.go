package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t04: Function Inlining Pass
//
// Replaces function calls with the body of the callee when the callee
// is small, non-recursive, and non-extern. This eliminates call overhead
// and enables further optimizations (constant folding across call
// boundaries, better register allocation).
// --------------------------------------------------------------------------

const (
	// DefaultInlineThreshold is the maximum number of non-NOP instructions
	// a callee may have to be eligible for inlining.
	DefaultInlineThreshold = 30
)

// InliningPass implements OptPass for function inlining.
type InliningPass struct {
	Threshold int // max callee instructions (0 = use default)
}

func (p *InliningPass) Name() string { return "inline" }

func (p *InliningPass) threshold() int {
	if p.Threshold > 0 {
		return p.Threshold
	}
	return DefaultInlineThreshold
}

// Run scans all functions for OpCall sites and inlines eligible callees.
func (p *InliningPass) Run(mod *air.AirModule) bool {
	if len(mod.Funcs) == 0 {
		return false
	}

	// Build function index: NameID → function index
	funcByName := make(map[uint32]int, len(mod.Funcs))
	for i := range mod.Funcs {
		funcByName[mod.Funcs[i].Name] = i
	}

	changed := false
	for fi := range mod.Funcs {
		caller := &mod.Funcs[fi]
		if inlineCalls(caller, mod, funcByName, p.threshold()) {
			changed = true
		}
	}
	return changed
}

// inlineCalls scans a single function for inlineable call sites.
func inlineCalls(caller *air.AirFunc, mod *air.AirModule, funcByName map[uint32]int, threshold int) bool {
	changed := false

	for i := 0; i < len(caller.Insts); i++ {
		inst := &caller.Insts[i]
		if inst.Opcode != air.OpCall {
			continue
		}

		// Src1 holds the callee NameID (or function index in extras)
		calleeNameID := inst.Src1
		calleeIdx, ok := funcByName[calleeNameID]
		if !ok {
			continue // unknown function (extern, function pointer)
		}

		callee := &mod.Funcs[calleeIdx]
		if !isInlineable(callee, caller, threshold) {
			continue
		}

		// Inline the callee at this call site
		inlineCallSite(caller, uint32(i), callee)
		changed = true

		// After inlining, the instruction list has changed.
		// Restart scan from the current position to avoid skipping.
		// (The inlined code replaces the OpCall, so we re-check the same index.)
	}
	return changed
}

// isInlineable checks whether a callee is eligible for inlining into the caller.
func isInlineable(callee *air.AirFunc, caller *air.AirFunc, threshold int) bool {
	// Don't inline recursive functions
	if callee.Name == caller.Name {
		return false
	}

	// Don't inline extern functions (no body)
	if callee.IsExtern {
		return false
	}

	// Don't inline async functions (state machine complexity)
	if callee.IsAsync {
		return false
	}

	// Cost check: count non-NOP instructions
	cost := costOf(callee)
	if cost > threshold {
		return false
	}

	// Check for self-recursion in the callee
	if hasSelfCall(callee) {
		return false
	}

	return true
}

// costOf counts the number of non-NOP instructions in a function.
func costOf(fn *air.AirFunc) int {
	count := 0
	for i := range fn.Insts {
		if fn.Insts[i].Opcode != air.OpNop {
			count++
		}
	}
	return count
}

// hasSelfCall checks if a function contains a call to itself (directly recursive).
func hasSelfCall(fn *air.AirFunc) bool {
	for i := range fn.Insts {
		if fn.Insts[i].Opcode == air.OpCall && fn.Insts[i].Src1 == fn.Name {
			return true
		}
	}
	return false
}

// inlineCallSite replaces an OpCall instruction with the cloned body of the callee.
func inlineCallSite(caller *air.AirFunc, callIdx uint32, callee *air.AirFunc) {
	// Determine register offset for cloning
	maxReg := findMaxReg(caller)

	// Clone callee instructions with remapped registers
	cloned := cloneInstructions(callee, maxReg)

	// Replace OpReturn in cloned code with OpNop (the return value is
	// captured by assigning callee's return reg to the call's dest reg)
	callInst := caller.Insts[callIdx]
	callDest := callInst.Dest

	for i := range cloned {
		if cloned[i].Opcode == air.OpReturn {
			if cloned[i].Src1 != 0 && callDest != 0 {
				// Replace return with a copy of the return value to the call dest
				cloned[i] = air.AirInst{
					Opcode: air.OpCopy,
					TypeID: callInst.TypeID,
					Dest:   callDest,
					Src1:   cloned[i].Src1,
				}
			} else {
				cloned[i] = air.AirInst{Opcode: air.OpNop}
			}
		}
	}

	// Replace the OpCall with the first cloned instruction, and insert the rest
	if len(cloned) == 0 {
		caller.Insts[callIdx] = air.AirInst{Opcode: air.OpNop}
		return
	}

	// Replace call with first cloned inst
	caller.Insts[callIdx] = cloned[0]

	// Insert remaining cloned instructions after the call site
	if len(cloned) > 1 {
		rest := cloned[1:]
		newInsts := make([]air.AirInst, 0, len(caller.Insts)+len(rest))
		newInsts = append(newInsts, caller.Insts[:callIdx+1]...)
		newInsts = append(newInsts, rest...)
		newInsts = append(newInsts, caller.Insts[callIdx+1:]...)
		caller.Insts = newInsts

		// Update block instruction indices for the current block
		updateBlockInstrs(caller, callIdx, uint32(len(rest)))
	}
}

// findMaxReg finds the highest register number used in a function.
func findMaxReg(fn *air.AirFunc) uint32 {
	maxReg := uint32(0)
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Dest > maxReg {
			maxReg = inst.Dest
		}
		if inst.Src1 > maxReg && !inst.Opcode.IsControl() {
			maxReg = inst.Src1
		}
		if inst.Src2 > maxReg && !inst.Opcode.IsControl() {
			maxReg = inst.Src2
		}
	}
	return maxReg
}

// cloneInstructions creates a deep copy of callee's instructions with
// register IDs offset by `offset` to avoid collisions with caller registers.
func cloneInstructions(callee *air.AirFunc, offset uint32) []air.AirInst {
	cloned := make([]air.AirInst, len(callee.Insts))
	for i := range callee.Insts {
		src := callee.Insts[i]
		if src.Opcode == air.OpNop {
			cloned[i] = src
			continue
		}

		// Remap registers (non-zero values only)
		if src.Dest != 0 {
			src.Dest += offset
		}

		// For non-control instructions, remap source registers
		if !src.Opcode.IsControl() && !src.Opcode.IsTerminator() {
			if src.Src1 != 0 {
				src.Src1 += offset
			}
			if src.Src2 != 0 {
				src.Src2 += offset
			}
		} else if src.Opcode == air.OpReturn {
			// Return value register needs remapping
			if src.Src1 != 0 {
				src.Src1 += offset
			}
		}

		// Don't remap constants
		if src.Opcode == air.OpIConst || src.Opcode == air.OpFConst {
			src.Src1 = callee.Insts[i].Src1 // restore original constant value
		}

		cloned[i] = src
	}
	return cloned
}

// updateBlockInstrs updates block instruction indices after inserting
// `count` new instructions after position `afterIdx`.
func updateBlockInstrs(fn *air.AirFunc, afterIdx uint32, count uint32) {
	for bi := range fn.Blocks {
		blk := &fn.Blocks[bi]
		for j := range blk.Instrs {
			if blk.Instrs[j] > afterIdx {
				blk.Instrs[j] += count
			}
		}
	}
}
