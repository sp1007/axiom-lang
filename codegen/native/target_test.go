package native_test

import (
	"runtime"
	"testing"

	"github.com/axiom-lang/axiom/codegen/native"
)

func TestParseTarget_X86_64_Linux(t *testing.T) {
	tgt, err := native.ParseTarget("x86_64-linux-gnu")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Arch != native.ArchX86_64 {
		t.Errorf("expected x86_64, got %s", tgt.Arch)
	}
	if tgt.OS != native.OSLinux {
		t.Errorf("expected linux, got %s", tgt.OS)
	}
	if tgt.ABI != native.ABISysV {
		t.Errorf("expected sysv, got %s", tgt.ABI)
	}
}

func TestParseTarget_X86_64_Windows(t *testing.T) {
	tgt, err := native.ParseTarget("x86_64-windows-win64")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Arch != native.ArchX86_64 {
		t.Errorf("expected x86_64, got %s", tgt.Arch)
	}
	if tgt.OS != native.OSWindows {
		t.Errorf("expected windows, got %s", tgt.OS)
	}
	if tgt.ABI != native.ABIWin64 {
		t.Errorf("expected win64, got %s", tgt.ABI)
	}
}

func TestParseTarget_ARM64_macOS(t *testing.T) {
	tgt, err := native.ParseTarget("aarch64-macos")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Arch != native.ArchARM64 {
		t.Errorf("expected aarch64, got %s", tgt.Arch)
	}
	if tgt.OS != native.OSmacOS {
		t.Errorf("expected macos, got %s", tgt.OS)
	}
	if tgt.ABI != native.ABIAAPCS64 {
		t.Errorf("expected aapcs64, got %s", tgt.ABI)
	}
}

func TestParseTarget_RISCV64_Linux(t *testing.T) {
	tgt, err := native.ParseTarget("riscv64-linux")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Arch != native.ArchRISCV64 {
		t.Errorf("expected riscv64, got %s", tgt.Arch)
	}
	if tgt.ABI != native.ABIRISCVpsABI {
		t.Errorf("expected riscv-psabi, got %s", tgt.ABI)
	}
}

func TestParseTarget_AMD64_Alias(t *testing.T) {
	tgt, err := native.ParseTarget("amd64-linux")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Arch != native.ArchX86_64 {
		t.Errorf("amd64 should map to x86_64, got %s", tgt.Arch)
	}
}

func TestParseTarget_ARM64_Alias(t *testing.T) {
	tgt, err := native.ParseTarget("arm64-linux")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Arch != native.ArchARM64 {
		t.Errorf("arm64 should map to aarch64, got %s", tgt.Arch)
	}
}

func TestParseTarget_Invalid(t *testing.T) {
	_, err := native.ParseTarget("invalid")
	if err == nil {
		t.Error("expected error for single-component triple")
	}

	_, err = native.ParseTarget("sparc-solaris")
	if err == nil {
		t.Error("expected error for unknown arch")
	}

	_, err = native.ParseTarget("x86_64-haiku")
	if err == nil {
		t.Error("expected error for unknown OS")
	}
}

func TestParseTarget_WindowsInferABI(t *testing.T) {
	tgt, err := native.ParseTarget("x86_64-windows")
	if err != nil {
		t.Fatal(err)
	}
	// Windows x86_64 should infer Win64 ABI
	if tgt.ABI != native.ABIWin64 {
		t.Errorf("expected win64 ABI for x86_64-windows, got %s", tgt.ABI)
	}
}

func TestHostTarget(t *testing.T) {
	tgt := native.HostTarget()

	switch runtime.GOARCH {
	case "amd64":
		if tgt.Arch != native.ArchX86_64 {
			t.Errorf("expected x86_64 on amd64, got %s", tgt.Arch)
		}
	case "arm64":
		if tgt.Arch != native.ArchARM64 {
			t.Errorf("expected aarch64 on arm64, got %s", tgt.Arch)
		}
	}

	switch runtime.GOOS {
	case "linux":
		if tgt.OS != native.OSLinux {
			t.Errorf("expected linux, got %s", tgt.OS)
		}
	case "windows":
		if tgt.OS != native.OSWindows {
			t.Errorf("expected windows, got %s", tgt.OS)
		}
	case "darwin":
		if tgt.OS != native.OSmacOS {
			t.Errorf("expected macos, got %s", tgt.OS)
		}
	}
}

func TestTargetProperties(t *testing.T) {
	tests := []struct {
		triple    string
		ptrSize   int
		intRegs   int
		floatRegs int
		argRegs   int
		binFmt    native.BinaryFmt
	}{
		{"x86_64-linux-gnu", 8, 16, 16, 6, native.BinELF},
		{"x86_64-windows-win64", 8, 16, 16, 4, native.BinPE},
		{"aarch64-macos", 8, 31, 32, 8, native.BinMachO},
		{"aarch64-linux", 8, 31, 32, 8, native.BinELF},
		{"riscv64-linux", 8, 32, 32, 8, native.BinELF},
	}

	for _, tt := range tests {
		t.Run(tt.triple, func(t *testing.T) {
			tgt, err := native.ParseTarget(tt.triple)
			if err != nil {
				t.Fatal(err)
			}
			if tgt.PointerSize() != tt.ptrSize {
				t.Errorf("PointerSize: got %d, want %d", tgt.PointerSize(), tt.ptrSize)
			}
			if tgt.IntRegCount() != tt.intRegs {
				t.Errorf("IntRegCount: got %d, want %d", tgt.IntRegCount(), tt.intRegs)
			}
			if tgt.FloatRegCount() != tt.floatRegs {
				t.Errorf("FloatRegCount: got %d, want %d", tgt.FloatRegCount(), tt.floatRegs)
			}
			if tgt.IntArgRegs() != tt.argRegs {
				t.Errorf("IntArgRegs: got %d, want %d", tgt.IntArgRegs(), tt.argRegs)
			}
			if tgt.BinaryFormat() != tt.binFmt {
				t.Errorf("BinaryFormat: got %s, want %s", tgt.BinaryFormat(), tt.binFmt)
			}
		})
	}
}

func TestTarget_Triple(t *testing.T) {
	tgt := native.Target{
		Arch: native.ArchX86_64,
		OS:   native.OSLinux,
		ABI:  native.ABISysV,
	}
	expected := "x86_64-linux-sysv"
	if tgt.Triple() != expected {
		t.Errorf("expected %q, got %q", expected, tgt.Triple())
	}
}

func TestTarget_String(t *testing.T) {
	tgt := native.Target{Arch: native.ArchARM64, OS: native.OSmacOS, ABI: native.ABIAAPCS64}
	if tgt.String() != "aarch64-macos-aapcs64" {
		t.Errorf("unexpected String(): %q", tgt.String())
	}
}
