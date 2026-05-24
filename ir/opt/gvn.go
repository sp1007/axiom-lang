package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// Global Value Numbering (GVN) Pass
//
// Identifies and eliminates redundant computations across basic blocks.
// An instruction is redundant if there is a previous equivalent computation
// with the same opcode, type, and operands.
//
// Since AXIOM's AIR uses a relaxed SSA form where mutable registers are modified
// via OpCopy, GVN only optimizes expressions whose operands are strictly
// immutable (defined exactly once: defCounts[reg] <= 1).
//
// When a redundant instruction is found, it is replaced with a copy instruction
// (`OpCopy`) pointing to the original value. The copy propagation pass will
// later propagate the original value and DCE will remove the redundant copy.
// --------------------------------------------------------------------------

type GVNPass struct{}

func (p *GVNPass) Name() string { return "gvn" }

func (p *GVNPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		if gvnFunc(&mod.Funcs[fi]) {
			changed = true
		}
	}
	return changed
}

type gvnKey struct {
	opcode air.Opcode
	typeID uint16
	src1   uint32
	src2   uint32
}

func gvnFunc(fn *air.AirFunc) bool {
	if len(fn.Insts) == 0 {
		return false
	}

	// Step 1: Count definitions of all virtual registers.
	// In AXIOM's SSA, registers defined exactly once (defCounts <= 1) are immutable.
	defCounts := make(map[uint32]int, len(fn.Insts))
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}
		if inst.Dest != 0 {
			defCounts[inst.Dest]++
		}
	}

	// Step 2: Global pre-order GVN traversal.
	// Since AIR is in SSA form, we can traverse instructions linearly.
	// Key: expression signature -> original destination register
	exprMap := make(map[gvnKey]uint32)
	changed := false

	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Opcode == air.OpNop {
			continue
		}

		// Only optimize instructions that define a destination register
		if inst.Dest == 0 {
			continue
		}

		// Check if opcode is eligible for GVN
		if !isGVNEligible(inst.Opcode) {
			continue
		}

		// Verify that all input operands are immutable SSA registers
		if !isOperandImmutable(inst.Src1, defCounts) || !isOperandImmutable(inst.Src2, defCounts) {
			continue
		}

		// Construct the expression key
		key := gvnKey{
			opcode: inst.Opcode,
			typeID: inst.TypeID,
			src1:   inst.Src1,
			src2:   inst.Src2,
		}

		if prevDest, exists := exprMap[key]; exists {
			// Redundant computation found! Replace with copy.
			inst.Opcode = air.OpCopy
			inst.Src1 = prevDest
			inst.Src2 = 0
			changed = true
		} else {
			// First time seeing this computation; register it.
			exprMap[key] = inst.Dest
		}
	}

	return changed
}

// isGVNEligible returns true if the opcode computes a deterministic value without side effects.
func isGVNEligible(op air.Opcode) bool {
	// Constants
	if op == air.OpIConst || op == air.OpFConst {
		return true
	}

	// Binary ALU operations
	if op.IsBinaryALU() {
		return true
	}

	// Unary operations & conversions
	switch op {
	case air.OpNeg, air.OpNot, air.OpIToF, air.OpFToI, air.OpZExt, air.OpSExt, air.OpTrunc, air.OpCast:
		return true
	case air.OpGEP, air.OpGetField:
		return true
	}

	return false
}

// isOperandImmutable returns true if the register is either unused (0) or immutable (defined exactly once).
func isOperandImmutable(reg uint32, defCounts map[uint32]int) bool {
	if reg == 0 {
		return true
	}
	return defCounts[reg] <= 1
}
