package x86

// --------------------------------------------------------------------------
// p11-t08: System V AMD64 ABI
// p11-t09: Win64 ABI
//
// Defines the ABI interface and two implementations: SysV (Linux/macOS)
// and Win64 (Windows). Each ABI specifies argument passing registers,
// callee-saved registers, return value handling, and calling conventions.
// --------------------------------------------------------------------------

// ABI defines the calling convention interface for a target.
type ABI interface {
	// IntArgRegs returns the registers for integer/pointer arguments.
	IntArgRegs() []PhysReg

	// FloatArgRegs returns the registers for float/double arguments.
	FloatArgRegs() []PhysReg

	// ReturnReg returns the register used for integer return values.
	ReturnReg() PhysReg

	// CalleeSavedRegs returns the list of callee-saved GPRs.
	CalleeSavedRegs() []PhysReg

	// CallerSavedRegs returns the list of caller-saved GPRs.
	CallerSavedRegs() []PhysReg

	// StackAlignment returns the required stack alignment in bytes.
	StackAlignment() int

	// ShadowSpace returns the required shadow/home space in bytes (Win64 only).
	ShadowSpace() int

	// Name returns the ABI name.
	Name() string
}

// NewABI creates the appropriate ABI implementation for the given ABI name.
// Accepts "win64" for Win64ABI, anything else defaults to SysVABI.
func NewABI(abiName string) ABI {
	switch abiName {
	case "win64":
		return &Win64ABI{}
	default:
		return &SysVABI{}
	}
}

// --------------------------------------------------------------------------
// System V AMD64 ABI (Linux, macOS, FreeBSD)
// --------------------------------------------------------------------------

// SysVABI implements the System V AMD64 calling convention.
type SysVABI struct{}

func (a *SysVABI) Name() string { return "sysv" }

func (a *SysVABI) IntArgRegs() []PhysReg {
	return []PhysReg{RDI, RSI, RDX, RCX, R8, R9}
}

func (a *SysVABI) FloatArgRegs() []PhysReg {
	return []PhysReg{XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7}
}

func (a *SysVABI) ReturnReg() PhysReg { return RAX }

func (a *SysVABI) CalleeSavedRegs() []PhysReg {
	return []PhysReg{RBX, R12, R13, R14, R15}
}

func (a *SysVABI) CallerSavedRegs() []PhysReg {
	return []PhysReg{RAX, RCX, RDX, RSI, RDI, R8, R9, R10, R11}
}

func (a *SysVABI) StackAlignment() int { return 16 }
func (a *SysVABI) ShadowSpace() int    { return 0 }

// --------------------------------------------------------------------------
// Win64 ABI (Windows x64)
// --------------------------------------------------------------------------

// Win64ABI implements the Windows x64 calling convention.
type Win64ABI struct{}

func (a *Win64ABI) Name() string { return "win64" }

func (a *Win64ABI) IntArgRegs() []PhysReg {
	return []PhysReg{RCX, RDX, R8, R9}
}

func (a *Win64ABI) FloatArgRegs() []PhysReg {
	return []PhysReg{XMM0, XMM1, XMM2, XMM3}
}

func (a *Win64ABI) ReturnReg() PhysReg { return RAX }

func (a *Win64ABI) CalleeSavedRegs() []PhysReg {
	return []PhysReg{RBX, RBP, RDI, RSI, R12, R13, R14, R15}
}

func (a *Win64ABI) CallerSavedRegs() []PhysReg {
	return []PhysReg{RAX, RCX, RDX, R8, R9, R10, R11}
}

func (a *Win64ABI) StackAlignment() int { return 16 }
func (a *Win64ABI) ShadowSpace() int    { return 32 } // mandatory 32-byte shadow space
