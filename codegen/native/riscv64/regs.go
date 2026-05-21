package riscv64

// --------------------------------------------------------------------------
// p13-t05: RISC-V 64-bit Register Definitions
//
// RV64I base integer register file: 32 GPRs (x0-x31) plus 32 FP regs.
// Uses the standard RISC-V psABI naming convention.
// --------------------------------------------------------------------------

// PhysReg represents a RISC-V 64-bit physical register.
type PhysReg uint8

// Integer registers (x0-x31) with ABI names
const (
	Zero PhysReg = iota // x0: hardwired zero
	RA                   // x1: return address
	SP                   // x2: stack pointer
	GP                   // x3: global pointer
	TP                   // x4: thread pointer
	T0                   // x5: temporary
	T1                   // x6: temporary
	T2                   // x7: temporary
	S0                   // x8 / FP: callee-saved / frame pointer
	S1                   // x9: callee-saved
	A0                   // x10: argument / return
	A1                   // x11: argument / return
	A2                   // x12: argument
	A3                   // x13: argument
	A4                   // x14: argument
	A5                   // x15: argument
	A6                   // x16: argument
	A7                   // x17: argument
	S2                   // x18: callee-saved
	S3                   // x19: callee-saved
	S4                   // x20: callee-saved
	S5                   // x21: callee-saved
	S6                   // x22: callee-saved
	S7                   // x23: callee-saved
	S8                   // x24: callee-saved
	S9                   // x25: callee-saved
	S10                  // x26: callee-saved
	S11                  // x27: callee-saved
	T3                   // x28: temporary
	T4                   // x29: temporary
	T5                   // x30: temporary
	T6                   // x31: temporary
)

// FP registers (f0-f31)
const (
	FT0  PhysReg = 0x20 + iota
	FT1
	FT2
	FT3
	FT4
	FT5
	FT6
	FT7
	FS0
	FS1
	FA0
	FA1
	FA2
	FA3
	FA4
	FA5
	FA6
	FA7
	FS2
	FS3
	FS4
	FS5
	FS6
	FS7
	FS8
	FS9
	FS10
	FS11
	FT8
	FT9
	FT10
	FT11
)

// RegNone represents no register.
const RegNone PhysReg = 0xFF

// IsGPR returns true if this is an integer register.
func (r PhysReg) IsGPR() bool { return r < 32 }

// IsFP returns true if this is a floating-point register.
func (r PhysReg) IsFP() bool { return r >= 0x20 && r < 0x40 }

// HWReg returns the 5-bit hardware register number.
func (r PhysReg) HWReg() uint8 {
	if r.IsFP() {
		return uint8(r) - 0x20
	}
	return uint8(r) & 0x1F
}

// String returns the ABI name of the register.
func (r PhysReg) String() string {
	if r.IsGPR() {
		return gprNames[r]
	}
	if r.IsFP() {
		hw := r.HWReg()
		if hw < 32 {
			return fpNames[hw]
		}
	}
	if r == RegNone {
		return "none"
	}
	return "???"
}

var gprNames = [32]string{
	"zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
	"s0", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
	"a6", "a7", "s2", "s3", "s4", "s5", "s6", "s7",
	"s8", "s9", "s10", "s11", "t3", "t4", "t5", "t6",
}

var fpNames = [32]string{
	"ft0", "ft1", "ft2", "ft3", "ft4", "ft5", "ft6", "ft7",
	"fs0", "fs1", "fa0", "fa1", "fa2", "fa3", "fa4", "fa5",
	"fa6", "fa7", "fs2", "fs3", "fs4", "fs5", "fs6", "fs7",
	"fs8", "fs9", "fs10", "fs11", "ft8", "ft9", "ft10", "ft11",
}

// AllocatableGPRs returns GPRs available for register allocation.
// Excludes x0 (zero), x1 (ra), x2 (sp), x3 (gp), x4 (tp).
func AllocatableGPRs() []PhysReg {
	return []PhysReg{
		T0, T1, T2,
		S0, S1,
		A0, A1, A2, A3, A4, A5, A6, A7,
		S2, S3, S4, S5, S6, S7, S8, S9, S10, S11,
		T3, T4, T5, T6,
	}
}
