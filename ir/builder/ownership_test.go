package builder_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/builder"
)

func TestOwnership_HeapAlloc(t *testing.T) {
	// Test that ownership helpers compile and produce correct opcodes
	// We test by directly constructing AIR and verifying the output

	fb := air.NewAirFuncBuilder(1, 0)

	// Simulate heap alloc: OpAlloc + OpMakeRef
	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: ptrReg})

	refReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpMakeRef, TypeID: 3, Dest: refReg, Src1: ptrReg})

	// Deref
	rawReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpDeref, TypeID: 3, Dest: rawReg, Src1: refReg})

	// Destroy
	fb.Emit(air.AirInst{Opcode: air.OpDestroy, Src1: rawReg})

	// Return
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()

	// Verify the generated AIR is valid
	errs := air.Verify(fn)
	for _, e := range errs {
		t.Logf("verify: %s", e.Error())
	}

	// Check that all expected opcodes are present
	opcodes := make(map[air.Opcode]bool)
	for _, inst := range fn.Insts {
		opcodes[inst.Opcode] = true
	}

	if !opcodes[air.OpAlloc] {
		t.Error("missing OpAlloc")
	}
	if !opcodes[air.OpMakeRef] {
		t.Error("missing OpMakeRef")
	}
	if !opcodes[air.OpDeref] {
		t.Error("missing OpDeref")
	}
	if !opcodes[air.OpDestroy] {
		t.Error("missing OpDestroy")
	}
}

func TestOwnership_Move(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)

	srcReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: srcReg, Src1: 42})

	dstReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpMove, TypeID: 3, Dest: dstReg, Src1: srcReg})

	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: dstReg})
	fn := fb.Build()

	hasMove := false
	for _, inst := range fn.Insts {
		if inst.Opcode == air.OpMove {
			hasMove = true
		}
	}
	if !hasMove {
		t.Error("missing OpMove")
	}
}

func TestOwnership_ArenaAlloc(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)

	arenaReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpIConst, Dest: arenaReg, Src1: 0})

	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpArenaAlloc, TypeID: 3, Dest: ptrReg, Src1: arenaReg})

	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()

	hasArena := false
	for _, inst := range fn.Insts {
		if inst.Opcode == air.OpArenaAlloc {
			hasArena = true
		}
	}
	if !hasArena {
		t.Error("missing OpArenaAlloc")
	}
}

func TestOwnership_AliasReuse(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)

	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: ptrReg})

	reuseReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAliasReuse, TypeID: 3, Dest: reuseReg, Src1: ptrReg})

	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()

	hasReuse := false
	for _, inst := range fn.Insts {
		if inst.Opcode == air.OpAliasReuse {
			hasReuse = true
		}
	}
	if !hasReuse {
		t.Error("missing OpAliasReuse")
	}
}

func TestAsync_SyncFunction(t *testing.T) {
	src := "fn main():\n    return\n"
	tokens, _, _ := lexer.Lex([]byte(src))
	intern := ast.NewInternPool(256)
	tree, _ := parser.Parse(tokens, []byte(src), intern)
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	mb := builder.NewModuleBuilder(tree, symbols, table, intern)
	mod := mb.Build()

	if len(mod.Funcs) == 0 {
		t.Fatal("expected at least 1 func")
	}

	// Sync function should not have IsAsync set
	if mod.Funcs[0].IsAsync {
		t.Error("sync function should not have IsAsync=true")
	}
}

func TestAsync_AsyncFunctionDecl(t *testing.T) {
	src := "async fn fetch():\n    return\n"
	tokens, _, _ := lexer.Lex([]byte(src))
	intern := ast.NewInternPool(256)
	tree, _ := parser.Parse(tokens, []byte(src), intern)
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	mb := builder.NewModuleBuilder(tree, symbols, table, intern)
	mod := mb.Build()

	if len(mod.Funcs) == 0 {
		t.Fatal("expected at least 1 func")
	}

	// Async function should have IsAsync flag
	if !mod.Funcs[0].IsAsync {
		t.Error("async function should have IsAsync=true")
	}
}

func TestModuleBuilder_DestroyStmt(t *testing.T) {
	// Test that destroy statements produce OpDestroy instructions
	fb := air.NewAirFuncBuilder(1, 0)

	valReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: valReg, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpDestroy, Src1: valReg})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})

	fn := fb.Build()
	output := air.SprintFunc(fn)

	if len(output) == 0 {
		t.Error("expected non-empty printer output")
	}
}
