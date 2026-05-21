package x86

// --------------------------------------------------------------------------
// p11-t07: x86-64 Stack Frame Layout
// p11-t06: Spill Code Generation
//
// Computes the stack frame layout including callee-saved registers,
// spill slots, local variables, and alignment padding. Generates
// prologue/epilogue MachInst sequences.
// --------------------------------------------------------------------------

// StackFrame describes the layout of a function's stack frame.
type StackFrame struct {
	CalleeSaved  []PhysReg // callee-saved registers to push/pop
	SpillSlots   int       // number of spill slots (8 bytes each)
	LocalBytes   int       // bytes for local variables
	AlignPadding int       // padding to maintain 16-byte alignment
	TotalSize    int       // total stack frame size (excluding pushed regs)
}

// ComputeFrame calculates the stack frame layout.
func ComputeFrame(calleeSaved []PhysReg, spillCount int, localBytes int) StackFrame {
	frame := StackFrame{
		CalleeSaved: calleeSaved,
		SpillSlots:  spillCount,
		LocalBytes:  localBytes,
	}

	// Stack space needed: spill slots + local variables
	needed := spillCount*8 + localBytes

	// After CALL, RSP is misaligned by 8 (return address pushed).
	// After PUSH RBP and pushing callee-saved regs, we need the total
	// (pushed + frame) to be 16-byte aligned.
	// pushed = return addr + RBP + len(calleeSaved)
	pushedBytes := (len(calleeSaved) + 2) * 8 // +2 for return address + RBP

	// Align total to 16 bytes
	total := needed
	if (pushedBytes+total)%16 != 0 {
		frame.AlignPadding = 16 - (pushedBytes+total)%16
		total += frame.AlignPadding
	}

	frame.TotalSize = total
	return frame
}

// SpillOffset returns the stack offset for a spill slot relative to RBP.
// Spill slots are at [RBP - 8*(slot+1)] (below saved registers).
func (f *StackFrame) SpillOffset(slotIdx int) int32 {
	return -int32((slotIdx + 1) * 8)
}

// EmitPrologue generates the function prologue MachInsts.
func EmitPrologue(frame *StackFrame) []MachInst {
	var insts []MachInst

	// PUSH RBP
	insts = append(insts, MachInst{Op: MachPush, Src1: Phys(RBP)})
	// MOV RBP, RSP
	insts = append(insts, MachInst{Op: MachMov, Dst: Phys(RBP), Src1: Phys(RSP)})

	// Push callee-saved registers
	for _, reg := range frame.CalleeSaved {
		insts = append(insts, MachInst{Op: MachPush, Src1: Phys(reg)})
	}

	// Allocate stack space
	if frame.TotalSize > 0 {
		insts = append(insts, MachInst{
			Op:   MachSub,
			Dst:  Phys(RSP),
			Src1: Imm(int64(frame.TotalSize)),
		})
	}

	return insts
}

// EmitEpilogue generates the function epilogue MachInsts.
func EmitEpilogue(frame *StackFrame) []MachInst {
	var insts []MachInst

	// Deallocate stack space
	if frame.TotalSize > 0 {
		insts = append(insts, MachInst{
			Op:   MachAdd,
			Dst:  Phys(RSP),
			Src1: Imm(int64(frame.TotalSize)),
		})
	}

	// Pop callee-saved registers (reverse order)
	for i := len(frame.CalleeSaved) - 1; i >= 0; i-- {
		insts = append(insts, MachInst{Op: MachPop, Dst: Phys(frame.CalleeSaved[i])})
	}

	// POP RBP
	insts = append(insts, MachInst{Op: MachPop, Dst: Phys(RBP)})
	// RET
	insts = append(insts, MachInst{Op: MachRet})

	return insts
}

// InsertSpillCode inserts load/store instructions for spilled registers.
// It scans all MachInsts and replaces VReg references to spilled regs
// with loads from / stores to their stack slots.
func InsertSpillCode(insts []MachInst, allocs map[uint32]RegAllocation, frame *StackFrame) []MachInst {
	var result []MachInst

	for _, inst := range insts {
		// Check if any operand references a spilled register
		// For spilled Src1: insert a load before the instruction
		if inst.Src1.Kind == OpndVReg {
			if alloc, ok := allocs[inst.Src1.VReg]; ok && alloc.Spilled {
				// Load from spill slot into a scratch register
				offset := frame.SpillOffset(alloc.SpillIdx)
				result = append(result, MachInst{
					Op:   MachLoad,
					Dst:  Phys(R11), // scratch register
					Src1: Phys(RBP),
					Src2: Imm(int64(offset)),
				})
				inst.Src1 = Phys(R11)
			}
		}

		if inst.Src2.Kind == OpndVReg {
			if alloc, ok := allocs[inst.Src2.VReg]; ok && alloc.Spilled {
				offset := frame.SpillOffset(alloc.SpillIdx)
				result = append(result, MachInst{
					Op:   MachLoad,
					Dst:  Phys(R10), // second scratch register
					Src1: Phys(RBP),
					Src2: Imm(int64(offset)),
				})
				inst.Src2 = Phys(R10)
			}
		}

		result = append(result, inst)

		// Check if Dst is spilled: insert a store after the instruction
		if inst.Dst.Kind == OpndVReg {
			if alloc, ok := allocs[inst.Dst.VReg]; ok && alloc.Spilled {
				offset := frame.SpillOffset(alloc.SpillIdx)
				result = append(result, MachInst{
					Op:   MachStore,
					Dst:  Phys(RBP),
					Src1: Phys(R11), // value was written to scratch
					Src2: Imm(int64(offset)),
				})
			}
		}
	}

	return result
}
