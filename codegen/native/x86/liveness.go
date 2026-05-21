package x86

// --------------------------------------------------------------------------
// p11-t04: Liveness Analysis
//
// Computes live intervals for each virtual register by scanning the
// linearized MachInst stream. A live interval records where a VReg
// is first defined and last used, which feeds into register allocation.
// --------------------------------------------------------------------------

// LiveInterval represents the lifetime of a virtual register.
type LiveInterval struct {
	VReg  uint32 // virtual register ID
	Start int    // first definition (instruction index)
	End   int    // last use (instruction index)
}

// ComputeLiveness builds live intervals from a flat list of MachInsts.
func ComputeLiveness(insts []MachInst) []LiveInterval {
	// Map VReg → interval
	intervals := make(map[uint32]*LiveInterval)

	for i, inst := range insts {
		// Process definitions (Dst)
		if inst.Dst.Kind == OpndVReg && inst.Dst.VReg != 0 {
			vreg := inst.Dst.VReg
			if _, ok := intervals[vreg]; !ok {
				intervals[vreg] = &LiveInterval{VReg: vreg, Start: i, End: i}
			}
			// Definition doesn't extend End — only uses do
		}

		// Process uses (Src1, Src2, and Dst when it's also a use)
		processUse := func(op MachOperand, idx int) {
			if op.Kind == OpndVReg && op.VReg != 0 {
				iv, ok := intervals[op.VReg]
				if !ok {
					// Use before def (function parameter or phi)
					intervals[op.VReg] = &LiveInterval{VReg: op.VReg, Start: 0, End: idx}
				} else if idx > iv.End {
					iv.End = idx
				}
			}
		}

		processUse(inst.Src1, i)
		processUse(inst.Src2, i)

		// For two-operand ops where Dst is also read (ADD dst, src)
		if isTwoOperandRead(inst.Op) {
			processUse(inst.Dst, i)
		}
	}

	// Collect and sort by Start
	result := make([]LiveInterval, 0, len(intervals))
	for _, iv := range intervals {
		result = append(result, *iv)
	}

	// Insertion sort by Start (small N, stable)
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j].Start > key.Start {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}

	return result
}

// isTwoOperandRead returns true if the instruction reads its Dst operand.
func isTwoOperandRead(op MachOpKind) bool {
	switch op {
	case MachAdd, MachSub, MachImul, MachAnd, MachOr, MachXor,
		MachShl, MachSar, MachNeg, MachNot:
		return true
	case MachCmp, MachTest, MachStore:
		return true
	default:
		return false
	}
}
