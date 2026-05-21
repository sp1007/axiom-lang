package air

import (
	"testing"
)

func TestNewAirFuncBuilder_CreatesEntryBlock(t *testing.T) {
	b := NewAirFuncBuilder(42, 1)
	if len(b.blocks) != 1 {
		t.Fatalf("expected 1 block after construction, got %d", len(b.blocks))
	}
	if !b.blocks[0].IsEntry {
		t.Error("block 0 should be marked IsEntry")
	}
	if b.CurrentBlock() != 0 {
		t.Errorf("CurrentBlock() = %d, want 0", b.CurrentBlock())
	}
}

func TestFreshReg_Unique(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r1 := b.FreshReg()
	r2 := b.FreshReg()
	r3 := b.FreshReg()
	if r1 == r2 || r2 == r3 || r1 == r3 {
		t.Errorf("FreshReg produced duplicate: %d, %d, %d", r1, r2, r3)
	}
	if r1 == 0 {
		t.Error("FreshReg should never return 0 (reserved for NoValue)")
	}
}

func TestEmit_AppendsToCurrentBlock(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	reg := b.FreshReg()
	idx := b.Emit(AirInst{Opcode: OpIConst, Dest: reg, Src1: 100})

	if idx != 0 {
		t.Errorf("first Emit index = %d, want 0", idx)
	}
	if len(b.blocks[0].Instrs) != 1 {
		t.Fatalf("block 0 should have 1 instr, got %d", len(b.blocks[0].Instrs))
	}
	if b.blocks[0].Instrs[0] != idx {
		t.Errorf("block 0 instr[0] = %d, want %d", b.blocks[0].Instrs[0], idx)
	}
}

func TestNewBlock_And_SwitchTo(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	b1 := b.NewBlock()
	if b1 != 1 {
		t.Errorf("second block ID = %d, want 1", b1)
	}
	b.SwitchTo(b1)
	if b.CurrentBlock() != b1 {
		t.Errorf("CurrentBlock() = %d, want %d", b.CurrentBlock(), b1)
	}
	idx := b.Emit(AirInst{Opcode: OpNop})
	if len(b.blocks[b1].Instrs) != 1 || b.blocks[b1].Instrs[0] != idx {
		t.Error("Emit did not append to switched block")
	}
}

func TestAddEdge(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	b1 := b.NewBlock()
	b.AddEdge(0, b1)

	if len(b.blocks[0].Succs) != 1 || b.blocks[0].Succs[0] != b1 {
		t.Errorf("block 0 succs = %v, want [%d]", b.blocks[0].Succs, b1)
	}
	if len(b.blocks[b1].Preds) != 1 || b.blocks[b1].Preds[0] != 0 {
		t.Errorf("block %d preds = %v, want [0]", b1, b.blocks[b1].Preds)
	}

	// Duplicate edges should not create duplicates.
	b.AddEdge(0, b1)
	if len(b.blocks[0].Succs) != 1 {
		t.Errorf("duplicate edge created: succs = %v", b.blocks[0].Succs)
	}
}

func TestBuild_MarksExitBlocks(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, Dest: r, Src1: 0})
	b.Emit(AirInst{Opcode: OpReturn, Src1: r})

	f := b.Build()
	if !f.Blocks[0].IsExit {
		t.Error("block 0 ending with OpReturn should be IsExit")
	}
	if f.Name != 1 {
		t.Errorf("AirFunc.Name = %d, want 1", f.Name)
	}
	if f.RetType != 1 {
		t.Errorf("AirFunc.RetType = %d, want 1", f.RetType)
	}
}

func TestEmitExtra(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	e0 := b.EmitExtra(10)
	e1 := b.EmitExtra(20)
	if e0 != 0 || e1 != 1 {
		t.Errorf("EmitExtra indices = (%d, %d), want (0, 1)", e0, e1)
	}
	f := b.Build()
	if len(f.Extras) != 2 || f.Extras[0] != 10 || f.Extras[1] != 20 {
		t.Errorf("Extras = %v, want [10 20]", f.Extras)
	}
}

// ---------------------------------------------------------------------------
// CFG traversal tests
// ---------------------------------------------------------------------------

// buildDiamondCFG creates:
//
//	entry(0)
//	├─→ left(1)
//	└─→ right(2)
//	     ├─→ merge(3)
//	     └─→ merge(3)
func buildDiamondCFG() *AirFunc {
	b := NewAirFuncBuilder(1, 1)
	left := b.NewBlock()   // 1
	right := b.NewBlock()  // 2
	merge := b.NewBlock()  // 3

	// entry → left, right
	cond := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, Dest: cond, Src1: 1})
	b.Emit(AirInst{Opcode: OpBranch, Src1: cond, Src2: left, Dest: right})
	b.AddEdge(0, left)
	b.AddEdge(0, right)

	// left → merge
	b.SwitchTo(left)
	b.Emit(AirInst{Opcode: OpJump, Src1: merge})
	b.AddEdge(left, merge)

	// right → merge
	b.SwitchTo(right)
	b.Emit(AirInst{Opcode: OpJump, Src1: merge})
	b.AddEdge(right, merge)

	// merge → return
	b.SwitchTo(merge)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, Dest: r, Src1: 0})
	b.Emit(AirInst{Opcode: OpReturn, Src1: r})

	return b.Build()
}

func TestPostOrder_Diamond(t *testing.T) {
	f := buildDiamondCFG()
	po := f.PostOrder()

	if len(po) != 4 {
		t.Fatalf("PostOrder length = %d, want 4", len(po))
	}
	// Entry (0) must be last in post-order.
	if po[len(po)-1] != 0 {
		t.Errorf("PostOrder last = %d, want 0 (entry)", po[len(po)-1])
	}
	// Merge (3) must come before left and right (it's the leaf in DFS).
	mergeIdx := indexOf(po, 3)
	leftIdx := indexOf(po, 1)
	rightIdx := indexOf(po, 2)
	if mergeIdx < 0 || leftIdx < 0 || rightIdx < 0 {
		t.Fatalf("missing block in PostOrder: %v", po)
	}
	// Merge must appear before left or right (post-order: children first).
	if mergeIdx > leftIdx && mergeIdx > rightIdx {
		t.Errorf("merge should appear before at least one of left/right in post-order: %v", po)
	}
}

func TestReversePostOrder_Diamond(t *testing.T) {
	f := buildDiamondCFG()
	rpo := f.ReversePostOrder()

	if len(rpo) != 4 {
		t.Fatalf("ReversePostOrder length = %d, want 4", len(rpo))
	}
	// Entry (0) must be first in reverse-post-order.
	if rpo[0] != 0 {
		t.Errorf("ReversePostOrder first = %d, want 0 (entry)", rpo[0])
	}
	// Merge (3) must be last.
	if rpo[len(rpo)-1] != 3 {
		t.Errorf("ReversePostOrder last = %d, want 3 (merge)", rpo[len(rpo)-1])
	}
}

func TestPostOrder_EmptyFunc(t *testing.T) {
	f := &AirFunc{}
	po := f.PostOrder()
	if po != nil {
		t.Errorf("PostOrder of empty func should be nil, got %v", po)
	}
}

func TestAirModule(t *testing.T) {
	f1 := NewAirFuncBuilder(1, 1)
	f1.Emit(AirInst{Opcode: OpReturn})
	f2 := NewAirFuncBuilder(2, 1)
	f2.Emit(AirInst{Opcode: OpReturn})

	mod := AirModule{
		Funcs: []AirFunc{*f1.Build(), *f2.Build()},
	}
	if len(mod.Funcs) != 2 {
		t.Errorf("AirModule.Funcs length = %d, want 2", len(mod.Funcs))
	}
}

// indexOf returns the position of v in s, or -1.
func indexOf(s []uint32, v uint32) int {
	for i, e := range s {
		if e == v {
			return i
		}
	}
	return -1
}
