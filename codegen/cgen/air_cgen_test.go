package cgen_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/ir/air"
)

func TestAirCGen_SimpleReturn(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()
	fn.Name = 1 // "main" in intern pool
	fn.RetType = 3

	pool := ast.NewInternPool(64)
	mainID := pool.InternString("main")
	fn.Name = mainID

	tt := types.NewTypeTable()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	if !strings.Contains(output, "r_1 = 42") {
		t.Errorf("expected 'r_1 = 42' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "return r_1") {
		t.Errorf("expected 'return r_1' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "main(") {
		t.Errorf("expected 'main(' in output, got:\n%s", output)
	}
}

func TestAirCGen_BinaryOps(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()

	pool := ast.NewInternPool(64)
	fn.Name = pool.InternString("add")
	fn.RetType = 3

	tt := types.NewTypeTable()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	if !strings.Contains(output, "r_3 = r_1 + r_2") {
		t.Errorf("expected 'r_3 = r_1 + r_2' in output, got:\n%s", output)
	}
}

func TestAirCGen_ControlFlow(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 11, Dest: 1, Src1: 1})
	thenBlock := fb.NewBlock()
	elseBlock := fb.NewBlock()
	fb.Emit(air.AirInst{Opcode: air.OpBranch, Src1: 1, Src2: thenBlock, Dest: elseBlock})
	fb.AddEdge(0, thenBlock)
	fb.AddEdge(0, elseBlock)

	fb.SwitchTo(thenBlock)
	fb.Emit(air.AirInst{Opcode: air.OpReturn})

	fb.SwitchTo(elseBlock)
	fb.Emit(air.AirInst{Opcode: air.OpReturn})

	fn := fb.Build()
	pool := ast.NewInternPool(64)
	fn.Name = pool.InternString("cf_test")

	tt := types.NewTypeTable()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	if !strings.Contains(output, "if (r_1)") {
		t.Errorf("expected branch in output, got:\n%s", output)
	}
	if !strings.Contains(output, "goto block_") {
		t.Errorf("expected goto in output, got:\n%s", output)
	}
	if !strings.Contains(output, "block_") {
		t.Errorf("expected block labels in output, got:\n%s", output)
	}
}

func TestAirCGen_ForwardDecl(t *testing.T) {
	pool := ast.NewInternPool(64)
	tt := types.NewTypeTable()

	fb1 := air.NewAirFuncBuilder(1, 0)
	fb1.Emit(air.AirInst{Opcode: air.OpReturn})
	fn1 := fb1.Build()
	fn1.Name = pool.InternString("foo")

	fb2 := air.NewAirFuncBuilder(2, 0)
	fb2.Emit(air.AirInst{Opcode: air.OpReturn})
	fn2 := fb2.Build()
	fn2.Name = pool.InternString("bar")

	mod := &air.AirModule{Funcs: []air.AirFunc{*fn1, *fn2}}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	// Forward declarations should appear before bodies
	fwdIdx := strings.Index(output, "_AX_foo(void);")
	bodyIdx := strings.Index(output, "_AX_foo(void) {")
	if fwdIdx < 0 || bodyIdx < 0 {
		t.Errorf("expected forward decl and body, got:\n%s", output)
	} else if fwdIdx >= bodyIdx {
		t.Error("forward declaration should appear before body")
	}
}

func TestAirCGen_VoidReturn(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()

	pool := ast.NewInternPool(64)
	fn.Name = pool.InternString("void_fn")

	tt := types.NewTypeTable()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	if !strings.Contains(output, "return;") {
		t.Errorf("expected 'return;' for void function, got:\n%s", output)
	}
}

func TestAirCGen_MemoryOps(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: 1})
	fb.Emit(air.AirInst{Opcode: air.OpFree, Src1: 1})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()

	pool := ast.NewInternPool(64)
	fn.Name = pool.InternString("mem_test")

	tt := types.NewTypeTable()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	if !strings.Contains(output, "malloc") {
		t.Errorf("expected malloc in output, got:\n%s", output)
	}
	if !strings.Contains(output, "free(r_1)") {
		t.Errorf("expected 'free(r_1)' in output, got:\n%s", output)
	}
}

func TestAirCGen_EmptyModule(t *testing.T) {
	pool := ast.NewInternPool(64)
	tt := types.NewTypeTable()
	mod := &air.AirModule{}

	gen := cgen.NewAirCGen(tt, pool)
	output := gen.Generate(mod)

	if !strings.Contains(output, "#include") {
		t.Error("expected includes even for empty module")
	}
}
