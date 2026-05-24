package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

func TestLoopUnroll_Basic(t *testing.T) {
	// Build a function representing a real compiler while loop:
	//   block_0 (entry):
	//     mut i = 1
	//     jump block_1 (condBlock)
	//   block_1 (condBlock):
	//     %limit = 3
	//     %cond = le %i, %limit
	//     branch %cond, block_2, block_3
	//   block_2 (bodyBlock):
	//     %i_new = iadd %i, 1
	//     %i = copy %i_new
	//     jump block_1 (condBlock)
	//   block_3 (exitBlock):
	//     ret
	fb := air.NewAirFuncBuilder(1, 0)
	
	condBlockID := fb.NewBlock()
	bodyBlockID := fb.NewBlock()
	exitBlockID := fb.NewBlock()
	
	// block_0 (entry)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 1}) // mut i = 1
	fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlockID})
	fb.AddEdge(0, condBlockID)
	
	// block_1 (condBlock)
	fb.SwitchTo(condBlockID)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 3}) // limit = 3
	fb.Emit(air.AirInst{Opcode: air.OpLe, TypeID: 11, Dest: 3, Src1: 1, Src2: 2}) // cond = i <= limit
	fb.Emit(air.AirInst{Opcode: air.OpBranch, Src1: 3, Src2: bodyBlockID, Dest: exitBlockID})
	fb.AddEdge(condBlockID, bodyBlockID)
	fb.AddEdge(condBlockID, exitBlockID)
	
	// block_2 (bodyBlock)
	fb.SwitchTo(bodyBlockID)
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 4, Src1: 1, Src2: 1}) // i_new = i + 1
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 1, Src1: 4})          // i = i_new
	fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlockID})
	fb.AddEdge(bodyBlockID, condBlockID) // backedge (condBlockID = 1 < bodyBlockID = 2)
	
	// block_3 (exitBlock)
	fb.SwitchTo(exitBlockID)
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	unroller := &opt.LoopUnrollPass{}
	changed := unroller.Run(mod)

	if !changed {
		t.Fatal("expected LoopUnrollPass to unroll static loop")
	}

	// Unrolled loop should have more blocks
	if len(mod.Funcs[0].Blocks) < 4 {
		t.Errorf("expected at least 4 blocks after unrolling, got %d", len(mod.Funcs[0].Blocks))
	}

	// The branch inside the cond block should be replaced with a jump
	condBlk := &mod.Funcs[0].Blocks[condBlockID]
	foundBranch := false
	foundJump := false
	for _, instIdx := range condBlk.Instrs {
		inst := &mod.Funcs[0].Insts[instIdx]
		if inst.Opcode == air.OpBranch {
			foundBranch = true
		}
		if inst.Opcode == air.OpJump {
			foundJump = true
		}
	}

	if foundBranch {
		t.Error("expected loop branch in cond block to be eliminated")
	}
	if !foundJump {
		t.Error("expected loop branch in cond block to be replaced with direct jump")
	}
}
