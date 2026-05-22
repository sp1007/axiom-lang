package native

import (
	"fmt"
	"runtime"
	"strings"
)

// --------------------------------------------------------------------------
// p11-t01: Target Triple Abstraction
//
// Describes the compilation target (architecture, OS, ABI) and enables
// cross-compilation. All backend code should consult the Target struct
// rather than hardcoding platform assumptions.
// --------------------------------------------------------------------------

// ArchKind identifies the CPU architecture.
type ArchKind uint8

const (
	ArchX86_64  ArchKind = iota // x86-64 / AMD64
	ArchARM64                   // AArch64 / ARM64
	ArchRISCV64                 // RISC-V 64-bit
	ArchWasm32                  // WebAssembly 32-bit
)

// String returns the canonical name for the architecture.
func (a ArchKind) String() string {
	switch a {
	case ArchX86_64:
		return "x86_64"
	case ArchARM64:
		return "aarch64"
	case ArchRISCV64:
		return "riscv64"
	case ArchWasm32:
		return "wasm32"
	default:
		return fmt.Sprintf("arch_%d", a)
	}
}

// OSKind identifies the operating system.
type OSKind uint8

const (
	OSLinux   OSKind = iota // Linux
	OSWindows               // Windows
	OSmacOS                 // macOS / Darwin
	OSWasm                  // WebAssembly environment
)

// String returns the canonical name for the OS.
func (o OSKind) String() string {
	switch o {
	case OSLinux:
		return "linux"
	case OSWindows:
		return "windows"
	case OSmacOS:
		return "macos"
	case OSWasm:
		return "wasm"
	default:
		return fmt.Sprintf("os_%d", o)
	}
}

// ABIKind identifies the calling convention / ABI.
type ABIKind uint8

const (
	ABISysV      ABIKind = iota // System V AMD64 ABI (Linux, macOS x86_64)
	ABIWin64                    // Windows x64 calling convention
	ABIAAPCS64                  // ARM64 AAPCS64 (all ARM64 platforms)
	ABIRISCVpsABI               // RISC-V psABI
	ABIWasm                     // WebAssembly calling convention
)

// String returns the canonical name for the ABI.
func (a ABIKind) String() string {
	switch a {
	case ABISysV:
		return "sysv"
	case ABIWin64:
		return "win64"
	case ABIAAPCS64:
		return "aapcs64"
	case ABIRISCVpsABI:
		return "riscv-psabi"
	case ABIWasm:
		return "wasm"
	default:
		return fmt.Sprintf("abi_%d", a)
	}
}

// BinaryFmt identifies the output binary format.
type BinaryFmt uint8

const (
	BinELF   BinaryFmt = iota // ELF (Linux, FreeBSD)
	BinPE                      // PE/COFF (Windows)
	BinMachO                   // Mach-O (macOS)
	BinWasm                    // WebAssembly Text / Binary
)

// String returns the format name.
func (f BinaryFmt) String() string {
	switch f {
	case BinELF:
		return "elf"
	case BinPE:
		return "pe"
	case BinMachO:
		return "macho"
	case BinWasm:
		return "wasm"
	default:
		return fmt.Sprintf("fmt_%d", f)
	}
}

// Target describes the complete compilation target.
type Target struct {
	Arch ArchKind
	OS   OSKind
	ABI  ABIKind
}

// PointerSize returns the pointer size in bytes for this target.
func (t Target) PointerSize() int {
	// All currently supported architectures are 64-bit
	return 8
}

// IntRegCount returns the number of general-purpose integer registers.
func (t Target) IntRegCount() int {
	switch t.Arch {
	case ArchX86_64:
		return 16 // RAX..R15
	case ArchARM64:
		return 31 // X0..X30
	case ArchRISCV64:
		return 32 // x0..x31
	default:
		return 16
	}
}

// FloatRegCount returns the number of floating-point / SIMD registers.
func (t Target) FloatRegCount() int {
	switch t.Arch {
	case ArchX86_64:
		return 16 // XMM0..XMM15
	case ArchARM64:
		return 32 // V0..V31
	case ArchRISCV64:
		return 32 // f0..f31
	default:
		return 16
	}
}

// CallerSavedIntRegs returns the number of caller-saved (volatile) integer registers.
func (t Target) CallerSavedIntRegs() int {
	switch t.ABI {
	case ABISysV:
		return 9 // RAX, RCX, RDX, RSI, RDI, R8-R11
	case ABIWin64:
		return 7 // RAX, RCX, RDX, R8-R11
	case ABIAAPCS64:
		return 18 // X0-X17
	case ABIRISCVpsABI:
		return 15 // t0-t6, a0-a7
	default:
		return 9
	}
}

// CalleeSavedIntRegs returns the number of callee-saved (non-volatile) integer registers.
func (t Target) CalleeSavedIntRegs() int {
	switch t.ABI {
	case ABISysV:
		return 5 // RBX, RBP, R12-R15 (RBP sometimes frame ptr)
	case ABIWin64:
		return 7 // RBX, RBP, RDI, RSI, R12-R15
	case ABIAAPCS64:
		return 10 // X19-X28
	case ABIRISCVpsABI:
		return 12 // s0-s11
	default:
		return 5
	}
}

// IntArgRegs returns the number of integer registers used for function arguments.
func (t Target) IntArgRegs() int {
	switch t.ABI {
	case ABISysV:
		return 6 // RDI, RSI, RDX, RCX, R8, R9
	case ABIWin64:
		return 4 // RCX, RDX, R8, R9
	case ABIAAPCS64:
		return 8 // X0-X7
	case ABIRISCVpsABI:
		return 8 // a0-a7
	default:
		return 6
	}
}

// BinaryFormat returns the binary object format for this target.
func (t Target) BinaryFormat() BinaryFmt {
	switch t.OS {
	case OSLinux:
		return BinELF
	case OSWindows:
		return BinPE
	case OSmacOS:
		return BinMachO
	case OSWasm:
		return BinWasm
	default:
		return BinELF
	}
}

// Triple returns the canonical triple string (e.g., "x86_64-linux-sysv").
func (t Target) Triple() string {
	return t.Arch.String() + "-" + t.OS.String() + "-" + t.ABI.String()
}

// String returns the Triple representation.
func (t Target) String() string {
	return t.Triple()
}

// HostTarget returns the Target representing the current host system.
func HostTarget() Target {
	t := Target{}

	// Detect architecture
	switch runtime.GOARCH {
	case "amd64":
		t.Arch = ArchX86_64
	case "arm64":
		t.Arch = ArchARM64
	case "riscv64":
		t.Arch = ArchRISCV64
	default:
		t.Arch = ArchX86_64 // fallback
	}

	// Detect OS
	switch runtime.GOOS {
	case "linux":
		t.OS = OSLinux
	case "windows":
		t.OS = OSWindows
	case "darwin":
		t.OS = OSmacOS
	default:
		t.OS = OSLinux // fallback
	}

	// Infer ABI from arch + OS
	t.ABI = inferABI(t.Arch, t.OS)
	return t
}

// inferABI determines the ABI from the architecture and OS combination.
func inferABI(arch ArchKind, os OSKind) ABIKind {
	switch arch {
	case ArchX86_64:
		if os == OSWindows {
			return ABIWin64
		}
		return ABISysV
	case ArchARM64:
		return ABIAAPCS64
	case ArchRISCV64:
		return ABIRISCVpsABI
	case ArchWasm32:
		return ABIWasm
	default:
		return ABISysV
	}
}

// ParseTarget parses a target triple string into a Target.
// Accepted formats:
//   - "x86_64-linux-gnu"   (arch-os-env, env mapped to ABI)
//   - "x86_64-linux-sysv"  (arch-os-abi)
//   - "aarch64-macos"      (arch-os, ABI inferred)
//   - "x86_64-windows"     (arch-os, ABI inferred)
//   - "wasm32"             (synthesized as wasm32-unknown-unknown)
func ParseTarget(triple string) (Target, error) {
	tripleNorm := strings.TrimSpace(triple)
	if tripleNorm == "wasm32" || tripleNorm == "wasm" || tripleNorm == "wasm32-unknown-unknown" {
		tripleNorm = "wasm32-unknown-unknown"
	}
	parts := strings.Split(tripleNorm, "-")
	if len(parts) < 2 {
		return Target{}, fmt.Errorf("invalid target triple %q: expected at least arch-os", triple)
	}

	t := Target{}

	// Parse architecture
	switch strings.ToLower(parts[0]) {
	case "x86_64", "amd64", "x86-64":
		t.Arch = ArchX86_64
	case "aarch64", "arm64":
		t.Arch = ArchARM64
	case "riscv64":
		t.Arch = ArchRISCV64
	case "wasm32", "wasm":
		t.Arch = ArchWasm32
	default:
		return Target{}, fmt.Errorf("unknown architecture %q in triple %q", parts[0], triple)
	}

	// Parse OS
	switch strings.ToLower(parts[1]) {
	case "linux":
		t.OS = OSLinux
	case "windows", "win32", "win64":
		t.OS = OSWindows
	case "macos", "darwin", "apple":
		t.OS = OSmacOS
	case "wasm", "unknown", "wasi":
		t.OS = OSWasm
	default:
		return Target{}, fmt.Errorf("unknown OS %q in triple %q", parts[1], triple)
	}

	// Parse ABI (optional third component)
	if len(parts) >= 3 {
		switch strings.ToLower(parts[2]) {
		case "sysv", "gnu", "musl":
			t.ABI = ABISysV
		case "win64", "msvc", "mingw":
			t.ABI = ABIWin64
		case "aapcs64":
			t.ABI = ABIAAPCS64
		case "riscv-psabi", "lp64d", "lp64":
			t.ABI = ABIRISCVpsABI
		case "wasm", "unknown":
			t.ABI = ABIWasm
		default:
			// Unknown env/ABI → infer from arch + OS
			t.ABI = inferABI(t.Arch, t.OS)
		}
	} else {
		t.ABI = inferABI(t.Arch, t.OS)
	}

	return t, nil
}
