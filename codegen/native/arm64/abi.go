package arm64

// --------------------------------------------------------------------------
// p13-t03: ARM64 ABI — AAPCS64 Calling Convention
//
// Implements the AAPCS64 (Procedure Call Standard for ARM 64-bit)
// calling convention. This defines argument passing, return values,
// and callee-saved register sets.
// --------------------------------------------------------------------------

// AAPCS64 implements the AAPCS64 calling convention.
type AAPCS64 struct{}

func (a *AAPCS64) Name() string { return "aapcs64" }

func (a *AAPCS64) IntArgRegs() []PhysReg {
	return []PhysReg{X0, X1, X2, X3, X4, X5, X6, X7}
}

func (a *AAPCS64) FloatArgRegs() []PhysReg {
	return []PhysReg{V0, V1, V2, V3, V4, V5, V6, V7}
}

func (a *AAPCS64) ReturnReg() PhysReg { return X0 }

func (a *AAPCS64) CalleeSavedRegs() []PhysReg {
	return []PhysReg{X19, X20, X21, X22, X23, X24, X25, X26, X27, X28}
}

func (a *AAPCS64) CallerSavedRegs() []PhysReg {
	return []PhysReg{X0, X1, X2, X3, X4, X5, X6, X7,
		X8, X9, X10, X11, X12, X13, X14, X15, X16, X17}
}

func (a *AAPCS64) StackAlignment() int { return 16 }
func (a *AAPCS64) ShadowSpace() int    { return 0 }

// --------------------------------------------------------------------------
// p13-t02: ARM64 Stack Frame Layout
//
// ARM64 uses STP/LDP for prologue/epilogue. Frame pointer is X29,
// link register is X30 (stored alongside FP).
// --------------------------------------------------------------------------

// StackFrame describes an ARM64 function's stack frame.
type StackFrame struct {
	CalleeSaved  []PhysReg
	SpillSlots   int
	LocalBytes   int
	AlignPadding int
	TotalSize    int
}

// ComputeFrame calculates the ARM64 stack frame layout.
func ComputeFrame(calleeSaved []PhysReg, spillCount int, localBytes int) StackFrame {
	frame := StackFrame{
		CalleeSaved: calleeSaved,
		SpillSlots:  spillCount,
		LocalBytes:  localBytes,
	}

	// Space needed: FP+LR (16 bytes) + callee-saved pairs + spills + locals
	needed := 16 + // FP + LR pair
		len(calleeSaved)*8 +
		spillCount*8 +
		localBytes

	// Align to 16 bytes
	total := needed
	if total%16 != 0 {
		frame.AlignPadding = 16 - total%16
		total += frame.AlignPadding
	}

	frame.TotalSize = total
	return frame
}

// SpillOffset returns the stack offset for a spill slot (relative to SP).
func (f *StackFrame) SpillOffset(slotIdx int) int16 {
	// Spill slots are after FP/LR and callee-saved registers
	base := 16 + len(f.CalleeSaved)*8
	return int16(base + slotIdx*8)
}

// EmitPrologue generates ARM64 function prologue.
func EmitPrologue(frame *StackFrame) []MachInst {
	var insts []MachInst

	if frame.TotalSize > 0 {
		// SUB SP, SP, #framesize
		insts = append(insts, MachInst{
			Op:   MachSub,
			Dst:  Phys(PhysReg(31)), // SP
			Src1: Phys(PhysReg(31)), // SP
			Src2: Imm(int64(frame.TotalSize)),
		})
	}

	// STP X29, X30, [SP, #0] (save FP and LR)
	insts = append(insts, MachInst{
		Op:   MachStp,
		Dst:  Phys(X29),
		Src1: Phys(X30),
		Src2: Phys(PhysReg(31)), // SP
	})

	// MOV X29, SP (set frame pointer)
	insts = append(insts, MachInst{
		Op:   MachMov,
		Dst:  Phys(X29),
		Src1: Phys(PhysReg(31)), // SP
	})

	// Save callee-saved registers in pairs
	for i := 0; i < len(frame.CalleeSaved); i += 2 {
		if i+1 < len(frame.CalleeSaved) {
			insts = append(insts, MachInst{
				Op:   MachStp,
				Dst:  Phys(frame.CalleeSaved[i]),
				Src1: Phys(frame.CalleeSaved[i+1]),
				Src2: Phys(PhysReg(31)), // SP
			})
		} else {
			insts = append(insts, MachInst{
				Op:   MachStr,
				Dst:  Phys(frame.CalleeSaved[i]),
				Src1: Phys(PhysReg(31)), // SP
			})
		}
	}

	return insts
}

// EmitEpilogue generates ARM64 function epilogue.
func EmitEpilogue(frame *StackFrame) []MachInst {
	var insts []MachInst

	// Restore callee-saved registers (reverse order, pairs)
	for i := len(frame.CalleeSaved) - 1; i >= 0; i -= 2 {
		if i-1 >= 0 {
			insts = append(insts, MachInst{
				Op:   MachLdp,
				Dst:  Phys(frame.CalleeSaved[i-1]),
				Src1: Phys(frame.CalleeSaved[i]),
				Src2: Phys(PhysReg(31)),
			})
		} else {
			insts = append(insts, MachInst{
				Op:   MachLdr,
				Dst:  Phys(frame.CalleeSaved[i]),
				Src1: Phys(PhysReg(31)),
			})
		}
	}

	// LDP X29, X30, [SP, #0] (restore FP and LR)
	insts = append(insts, MachInst{
		Op:   MachLdp,
		Dst:  Phys(X29),
		Src1: Phys(X30),
		Src2: Phys(PhysReg(31)),
	})

	// ADD SP, SP, #framesize
	if frame.TotalSize > 0 {
		insts = append(insts, MachInst{
			Op:   MachAdd,
			Dst:  Phys(PhysReg(31)),
			Src1: Phys(PhysReg(31)),
			Src2: Imm(int64(frame.TotalSize)),
		})
	}

	// RET
	insts = append(insts, MachInst{Op: MachRet})

	return insts
}
