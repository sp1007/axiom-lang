package wasm_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/wasm"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
)

func TestWasm_SimpleAdd(t *testing.T) {
	// Build an AIR function:
	// fn add(a: i32, b: i32) -> i32 {
	//     return a + b
	// }
	// In AIR, parameters are implicit starting values in the entry block,
	// and registerParams emits copies from param slots:
	//   %1 = copy 1 (a)
	//   %2 = copy 2 (b)
	//   %3 = iadd %1, %2
	//   ret %3
	fb := air.NewAirFuncBuilder(1, 3) // name ID 1, return type i32 (3)
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 1, Src1: 1})
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 2, Src1: 2})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	fn.Params = []uint32{3, 3} // two i32 params

	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	intern := ast.NewInternPool(256)
	ttable := types.NewTypeTable()

	backend := wasm.NewWasmBackend()
	backend.Pool = intern
	backend.Table = ttable

	wat, err := backend.Compile(mod)
	if err != nil {
		t.Fatalf("failed to compile to WAT: %v", err)
	}

	// Verify WAT output structures
	if !strings.Contains(wat, "(module") {
		t.Error("expected module declaration in WAT")
	}
	if !strings.Contains(wat, "(param $p1 i32) (param $p2 i32)") {
		t.Error("expected parameter declarations in WAT")
	}
	if !strings.Contains(wat, "(result i32)") {
		t.Error("expected result declaration in WAT")
	}
	if !strings.Contains(wat, "i32.add") {
		t.Error("expected i32.add operation in WAT")
	}
	if !strings.Contains(wat, "br_table") {
		t.Error("expected block dispatcher state-machine br_table in WAT")
	}
}

func TestWasm_ConditionBranch(t *testing.T) {
	// fn is_positive(x: i32) -> i32 {
	//     if x > 0 { return 1 } else { return 0 }
	// }
	// AIR:
	//   block_0: (entry)
	//     %1 = copy 1
	//     %2 = iconst 0
	//     %3 = gt %1, %2
	//     branch %3 block_1 block_2
	//   block_1:
	//     %4 = iconst 1
	//     ret %4
	//   block_2:
	//     %5 = iconst 0
	//     ret %5
	fb := air.NewAirFuncBuilder(1, 3)
	// Entry block (always ID 0)
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 1, Src1: 1})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 0})
	fb.Emit(air.AirInst{Opcode: air.OpGt, TypeID: 11, Dest: 3, Src1: 1, Src2: 2}) // Bool result

	b1 := fb.NewBlock()
	b2 := fb.NewBlock()

	fb.Emit(air.AirInst{Opcode: air.OpBranch, Src1: 3, Src2: b1, Dest: b2})
	fb.AddEdge(0, b1)
	fb.AddEdge(0, b2)

	// Block 1
	fb.SwitchTo(b1)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 4, Src1: 1})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 4})

	// Block 2
	fb.SwitchTo(b2)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 5, Src1: 0})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 5})

	fn := fb.Build()
	fn.Params = []uint32{3}

	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	backend := wasm.NewWasmBackend()
	wat, err := backend.Compile(mod)
	if err != nil {
		t.Fatalf("failed to compile conditional branch: %v", err)
	}

	if !strings.Contains(wat, "br_table $b_0 $b_1 $b_2") {
		t.Error("expected br_table dispatcher mapping all blocks")
	}
	if !strings.Contains(wat, "(if") {
		t.Error("expected conditional check with Wasm if instruction")
	}
	if !strings.Contains(wat, "i32.gt_s") {
		t.Error("expected i32.gt_s comparison instruction")
	}
}
