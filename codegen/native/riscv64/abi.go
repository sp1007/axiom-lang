package riscv64

// --------------------------------------------------------------------------
// p13-t06: RISC-V psABI Calling Convention
//
// Implements the RISC-V LP64D calling convention (ILP32D for 32-bit).
// Arguments in a0-a7, return in a0/a1, callee-saved s0-s11.
// --------------------------------------------------------------------------

// RV64ABI implements the RISC-V LP64D calling convention.
type RV64ABI struct{}

func (a *RV64ABI) Name() string { return "lp64d" }

func (a *RV64ABI) IntArgRegs() []PhysReg {
	return []PhysReg{A0, A1, A2, A3, A4, A5, A6, A7}
}

func (a *RV64ABI) FloatArgRegs() []PhysReg {
	return []PhysReg{FA0, FA1, FA2, FA3, FA4, FA5, FA6, FA7}
}

func (a *RV64ABI) ReturnReg() PhysReg { return A0 }

func (a *RV64ABI) CalleeSavedRegs() []PhysReg {
	return []PhysReg{S0, S1, S2, S3, S4, S5, S6, S7, S8, S9, S10, S11}
}

func (a *RV64ABI) CallerSavedRegs() []PhysReg {
	return []PhysReg{T0, T1, T2, T3, T4, T5, T6, A0, A1, A2, A3, A4, A5, A6, A7}
}

func (a *RV64ABI) StackAlignment() int { return 16 }
func (a *RV64ABI) ShadowSpace() int    { return 0 }

// --------------------------------------------------------------------------
// RISC-V Stack Frame
//
// RISC-V frame:
//   [saved RA]
//   [saved S0 (FP)]
//   [callee-saved regs]
//   [spill slots]
//   [locals]
// --------------------------------------------------------------------------

// StackFrame describes a RISC-V function's stack frame.
type StackFrame struct {
	CalleeSaved  []PhysReg
	SpillSlots   int
	LocalBytes   int
	AlignPadding int
	TotalSize    int
}

// ComputeFrame calculates the RISC-V stack frame layout.
func ComputeFrame(calleeSaved []PhysReg, spillCount int, localBytes int) StackFrame {
	frame := StackFrame{
		CalleeSaved: calleeSaved,
		SpillSlots:  spillCount,
		LocalBytes:  localBytes,
	}

	// Space: RA + callee-saved + spills + locals
	needed := 8 + // RA
		len(calleeSaved)*8 +
		spillCount*8 +
		localBytes

	total := needed
	if total%16 != 0 {
		frame.AlignPadding = 16 - total%16
		total += frame.AlignPadding
	}

	frame.TotalSize = total
	return frame
}

// SpillOffset returns the stack offset for a spill slot.
func (f *StackFrame) SpillOffset(slotIdx int) int16 {
	base := 8 + len(f.CalleeSaved)*8 // after RA + callee-saved
	return int16(base + slotIdx*8)
}
