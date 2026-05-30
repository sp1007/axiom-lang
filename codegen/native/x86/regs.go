package x86

// --------------------------------------------------------------------------
// p11-t02: x86-64 Register Definitions
//
// Defines the physical register file for x86-64, including GPR encoding
// (3-bit reg field + REX.B extension), XMM registers, and register
// classification (caller-saved, callee-saved, argument registers).
// --------------------------------------------------------------------------

// PhysReg represents a physical x86-64 register.
// The low 4 bits encode the hardware register number (0-15).
// Bit 4 indicates XMM (float) vs GPR (int) class.
type PhysReg uint8

// GPR registers (64-bit names)
const (
	RAX PhysReg = 0
	RCX PhysReg = 1
	RDX PhysReg = 2
	RBX PhysReg = 3
	RSP PhysReg = 4
	RBP PhysReg = 5
	RSI PhysReg = 6
	RDI PhysReg = 7
	R8  PhysReg = 8
	R9  PhysReg = 9
	R10 PhysReg = 10
	R11 PhysReg = 11
	R12 PhysReg = 12
	R13 PhysReg = 13
	R14 PhysReg = 14
	R15 PhysReg = 15
)

// XMM registers (SSE/AVX)
const (
	XMM0  PhysReg = 0x10 + iota // 16
	XMM1                        // 17
	XMM2                        // 18
	XMM3                        // 19
	XMM4                        // 20
	XMM5                        // 21
	XMM6                        // 22
	XMM7                        // 23
	XMM8                        // 24
	XMM9                        // 25
	XMM10                       // 26
	XMM11                       // 27
	XMM12                       // 28
	XMM13                       // 29
	XMM14                       // 30
	XMM15                       // 31
)

// RegNone represents no register / invalid register.
const RegNone PhysReg = 0xFF

// IsGPR returns true if this is a general-purpose register.
func (r PhysReg) IsGPR() bool { return r < 16 }

// IsXMM returns true if this is an XMM (SSE/AVX) register.
func (r PhysReg) IsXMM() bool { return r >= 0x10 && r < 0x20 }

// HWReg returns the 4-bit hardware register number (0-15).
func (r PhysReg) HWReg() uint8 {
	if r.IsXMM() {
		return uint8(r) - 0x10
	}
	return uint8(r) & 0x0F
}

// RegField returns the 3-bit register encoding for ModRM/SIB.
func (r PhysReg) RegField() uint8 {
	return r.HWReg() & 0x07
}

// NeedsREX returns true if this register requires a REX prefix
// (register number >= 8).
func (r PhysReg) NeedsREX() bool {
	return r.HWReg() >= 8
}

// String returns the register name.
func (r PhysReg) String() string {
	if r.IsGPR() {
		return gprNames[r.HWReg()]
	}
	if r.IsXMM() {
		return xmmNames[r.HWReg()]
	}
	if r == RegNone {
		return "none"
	}
	return "???"
}

var gprNames = [16]string{
	"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi",
	"r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15",
}

var xmmNames = [16]string{
	"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7",
	"xmm8", "xmm9", "xmm10", "xmm11", "xmm12", "xmm13", "xmm14", "xmm15",
}

// --------------------------------------------------------------------------
// Register Classification (System V AMD64 ABI)
// --------------------------------------------------------------------------

// SysVIntArgRegs are the registers used for integer arguments (in order).
var SysVIntArgRegs = [6]PhysReg{RDI, RSI, RDX, RCX, R8, R9}

// SysVCallerSaved are caller-saved (volatile) GPRs.
var SysVCallerSaved = [9]PhysReg{RAX, RCX, RDX, RSI, RDI, R8, R9, R10, R11}

// SysVCalleeSaved are callee-saved (non-volatile) GPRs.
var SysVCalleeSaved = [5]PhysReg{RBX, R12, R13, R14, R15}

// Win64IntArgRegs are the registers used for integer arguments (in order).
var Win64IntArgRegs = [4]PhysReg{RCX, RDX, R8, R9}

// Win64CalleeSaved are callee-saved (non-volatile) GPRs.
var Win64CalleeSaved = [7]PhysReg{RBX, RBP, RDI, RSI, R12, R13, R14}

// AllocatableGPRs returns the GPRs available for register allocation
// (excludes RSP and RBP which are reserved for stack/frame, and R10, R11 which are scratch).
func AllocatableGPRs() []PhysReg {
	return []PhysReg{
		RAX, RCX, RDX, RBX, RSI, RDI,
		R8, R9, R12, R13, R14, R15,
	}
}

// AllocatableXMMs returns the XMM registers available for register allocation.
// We use XMM8-XMM15 as allocatable float registers.
func AllocatableXMMs() []PhysReg {
	return []PhysReg{
		XMM8, XMM9, XMM10, XMM11,
		XMM12, XMM13, XMM14, XMM15,
	}
}

