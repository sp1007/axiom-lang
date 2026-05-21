package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

func TestComptime_Arithmetic(t *testing.T) {
	// fn compute() -> i32: return 3 + 4
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 3})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 4})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(fn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != 7 {
		t.Errorf("expected 7, got %d", result.IVal)
	}
}

func TestComptime_Multiplication(t *testing.T) {
	// fn compute() -> i32: return 6 * 7
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 6})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 7})
	fb.Emit(air.AirInst{Opcode: air.OpIMul, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(fn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != 42 {
		t.Errorf("expected 42, got %d", result.IVal)
	}
}

func TestComptime_DivByZero(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 0})
	fb.Emit(air.AirInst{Opcode: air.OpIDiv, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	_, err := interp.Interpret(fn, nil)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
	if err.Error() != "#run: division by zero" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestComptime_Comparison(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 11)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpLt, TypeID: 11, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(fn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != 1 {
		t.Errorf("expected 1 (true), got %d", result.IVal)
	}
}

func TestComptime_Negation(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpNeg, TypeID: 3, Dest: 2, Src1: 1})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 2})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(fn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != -42 {
		t.Errorf("expected -42, got %d", result.IVal)
	}
}

func TestComptime_AllocRejected(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: 1})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	_, err := interp.Interpret(fn, nil)
	if err == nil {
		t.Fatal("expected error for alloc in comptime")
	}
}

func TestComptime_VoidReturn(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(fn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != 0 {
		t.Errorf("expected 0 for void return, got %d", result.IVal)
	}
}

func TestComptime_WithArgs(t *testing.T) {
	// fn add(a, b) -> i32: return a + b
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	args := []opt.Value{
		{TypeID: 3, IVal: 30},
		{TypeID: 3, IVal: 12},
	}
	result, err := interp.Interpret(fn, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != 42 {
		t.Errorf("expected 42, got %d", result.IVal)
	}
}

func TestComptime_BitwiseOps(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 0xFF})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 0x0F})
	fb.Emit(air.AirInst{Opcode: air.OpAnd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(fn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IVal != 0x0F {
		t.Errorf("expected 0x0F, got %d", result.IVal)
	}
}

func TestComptimePass_NoOp(t *testing.T) {
	mod := &air.AirModule{}
	pass := &opt.ComptimePass{}
	changed := pass.Run(mod)
	if changed {
		t.Error("comptime pass should be no-op without #run markers")
	}
}
