package x86_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/native/x86"
	"github.com/axiom-lang/axiom/ir/air"
)

func TestAsmEmitter_EmitFunction_NASM(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()

	abi := x86.NewABI("sysv")
	machInsts := x86.Select(fn, abi, nil)
	intervals := x86.ComputeLiveness(machInsts)
	availRegs := []x86.PhysReg{x86.RAX, x86.RCX}
	allocResult := x86.GraphColoringAlloc(intervals, availRegs)
	
	used := []x86.PhysReg{x86.RBX}
	frame := x86.ComputeFrame(used, allocResult.SpillCount, 0)
	machInsts = x86.InsertSpillCode(machInsts, allocResult.Allocs, &frame)

	// Emit NASM assembly
	emitter := x86.NewAsmEmitter(allocResult.Allocs, "nasm")
	asm := emitter.EmitFunction("main", machInsts, &frame, func(id uint32) string {
		return "external_func"
	})

	if !strings.Contains(asm, "main:") {
		t.Error("expected function label main:")
	}
	if !strings.Contains(asm, "mov") {
		t.Error("expected mov instruction")
	}
	if !strings.Contains(asm, "ret") {
		t.Error("expected ret instruction")
	}
}

func TestAsmEmitter_EmitFunction_FASM(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 100})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()

	abi := x86.NewABI("sysv")
	machInsts := x86.Select(fn, abi, nil)
	intervals := x86.ComputeLiveness(machInsts)
	availRegs := []x86.PhysReg{x86.RAX, x86.RCX}
	allocResult := x86.GraphColoringAlloc(intervals, availRegs)
	
	used := []x86.PhysReg{x86.RBX}
	frame := x86.ComputeFrame(used, allocResult.SpillCount, 0)
	machInsts = x86.InsertSpillCode(machInsts, allocResult.Allocs, &frame)

	// Emit FASM assembly
	emitter := x86.NewAsmEmitter(allocResult.Allocs, "fasm")
	asm := emitter.EmitFunction("main", machInsts, &frame, func(id uint32) string {
		return "external_func"
	})

	if !strings.Contains(asm, "main:") {
		t.Error("expected function label main:")
	}
	if !strings.Contains(asm, "mov") {
		t.Error("expected mov instruction")
	}
}

func TestAsmEmitter_EmitFunction_WinAsm(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()

	abi := x86.NewABI("sysv")
	machInsts := x86.Select(fn, abi, nil)
	intervals := x86.ComputeLiveness(machInsts)
	availRegs := []x86.PhysReg{x86.RAX, x86.RCX}
	allocResult := x86.GraphColoringAlloc(intervals, availRegs)
	
	used := []x86.PhysReg{x86.RBX}
	frame := x86.ComputeFrame(used, allocResult.SpillCount, 0)
	machInsts = x86.InsertSpillCode(machInsts, allocResult.Allocs, &frame)

	// Emit WinAsm / MASM assembly
	emitter := x86.NewAsmEmitter(allocResult.Allocs, "winasm")
	asm := emitter.EmitFunction("main", machInsts, &frame, func(id uint32) string {
		return "external_func"
	})

	if !strings.Contains(asm, "main PROC") {
		t.Error("expected PROC declaration in MASM")
	}
	if !strings.Contains(asm, "main ENDP") {
		t.Error("expected ENDP declaration in MASM")
	}
	if !strings.Contains(asm, "mov") {
		t.Error("expected mov instruction")
	}
}
