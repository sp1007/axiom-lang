package arm64

// --------------------------------------------------------------------------
// p13-t01: ARM64 Register Definitions
//
// Defines the AArch64 physical register file: 31 general-purpose registers
// (X0-X30), SP, 32 SIMD/FP registers (V0-V31), and register classification
// for AAPCS64 calling convention.
// --------------------------------------------------------------------------

// PhysReg represents an ARM64 physical register.
type PhysReg uint8

// General-purpose registers (64-bit X registers)
const (
	X0  PhysReg = iota // argument/return
	X1                  // argument
	X2                  // argument
	X3                  // argument
	X4                  // argument
	X5                  // argument
	X6                  // argument
	X7                  // argument
	X8                  // indirect result
	X9                  // caller-saved temp
	X10                 // caller-saved temp
	X11                 // caller-saved temp
	X12                 // caller-saved temp
	X13                 // caller-saved temp
	X14                 // caller-saved temp
	X15                 // caller-saved temp
	X16                 // IP0 (intra-procedure scratch)
	X17                 // IP1 (intra-procedure scratch)
	X18                 // platform register (reserved)
	X19                 // callee-saved
	X20                 // callee-saved
	X21                 // callee-saved
	X22                 // callee-saved
	X23                 // callee-saved
	X24                 // callee-saved
	X25                 // callee-saved
	X26                 // callee-saved
	X27                 // callee-saved
	X28                 // callee-saved
	X29                 // FP (frame pointer)
	X30                 // LR (link register)
)

// SP is the stack pointer (encoded as register 31 in some instructions).
const SP PhysReg = 31

// SIMD/FP registers
const (
	V0  PhysReg = 0x20 + iota
	V1
	V2
	V3
	V4
	V5
	V6
	V7
	V8
	V9
	V10
	V11
	V12
	V13
	V14
	V15
	V16
	V17
	V18
	V19
	V20
	V21
	V22
	V23
	V24
	V25
	V26
	V27
	V28
	V29
	V30
	V31
)

// RegNone represents no register / invalid register.
const RegNone PhysReg = 0xFF

// IsGPR returns true if this is a general-purpose register.
func (r PhysReg) IsGPR() bool { return r < 32 }

// IsVec returns true if this is a SIMD/FP register.
func (r PhysReg) IsVec() bool { return r >= 0x20 && r < 0x40 }

// HWReg returns the 5-bit hardware register number (0-30, 31=SP/ZR).
func (r PhysReg) HWReg() uint8 {
	if r.IsVec() {
		return uint8(r) - 0x20
	}
	return uint8(r) & 0x1F
}

// String returns the register name.
func (r PhysReg) String() string {
	if r.IsGPR() {
		return gprNames[r.HWReg()]
	}
	if r.IsVec() {
		hw := r.HWReg()
		if hw < 32 {
			return vecNames[hw]
		}
	}
	if r == RegNone {
		return "none"
	}
	return "???"
}

var gprNames = [32]string{
	"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7",
	"x8", "x9", "x10", "x11", "x12", "x13", "x14", "x15",
	"x16", "x17", "x18", "x19", "x20", "x21", "x22", "x23",
	"x24", "x25", "x26", "x27", "x28", "x29", "x30", "sp",
}

var vecNames = [32]string{
	"v0", "v1", "v2", "v3", "v4", "v5", "v6", "v7",
	"v8", "v9", "v10", "v11", "v12", "v13", "v14", "v15",
	"v16", "v17", "v18", "v19", "v20", "v21", "v22", "v23",
	"v24", "v25", "v26", "v27", "v28", "v29", "v30", "v31",
}

// --------------------------------------------------------------------------
// Register Classification (AAPCS64)
// --------------------------------------------------------------------------

// AAPCS64IntArgRegs: X0-X7 for integer arguments.
var AAPCS64IntArgRegs = [8]PhysReg{X0, X1, X2, X3, X4, X5, X6, X7}

// AAPCS64CallerSaved: X0-X18 are caller-saved (volatile).
var AAPCS64CallerSaved = [18]PhysReg{
	X0, X1, X2, X3, X4, X5, X6, X7,
	X8, X9, X10, X11, X12, X13, X14, X15,
	X16, X17,
}

// AAPCS64CalleeSaved: X19-X28 are callee-saved (non-volatile).
var AAPCS64CalleeSaved = [10]PhysReg{
	X19, X20, X21, X22, X23, X24, X25, X26, X27, X28,
}

// AllocatableGPRs returns the GPRs available for register allocation.
// Excludes X18 (platform), X29 (FP), X30 (LR), SP.
func AllocatableGPRs() []PhysReg {
	return []PhysReg{
		X0, X1, X2, X3, X4, X5, X6, X7,
		X8, X9, X10, X11, X12, X13, X14, X15,
		X19, X20, X21, X22, X23, X24, X25, X26, X27, X28,
	}
}
