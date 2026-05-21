package cgen_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// --- UnsafeDeref Tests ---

func TestUnsafeDeref_PrimitiveType(t *testing.T) {
	table, intern, queue := helper()

	got := cgen.UnsafeDeref("myref", types.TypeI32, table, intern, queue)
	if got != "(*((ax_i32*)(myref).ptr))" {
		t.Errorf("UnsafeDeref(i32) = %q", got)
	}
}

func TestUnsafeDeref_StructType(t *testing.T) {
	table, intern, queue := helper()
	nodeName := intern.InternString("Node")
	nodeID := table.RegisterStruct(nodeName, nil, nil)

	got := cgen.UnsafeDeref("node_ref", nodeID, table, intern, queue)
	if got != "(*((struct ax_Node*)(node_ref).ptr))" {
		t.Errorf("UnsafeDeref(Node) = %q", got)
	}
}

// --- UncheckedIndex Tests ---

func TestUncheckedIndex(t *testing.T) {
	got := cgen.UncheckedIndex("arr", "i")
	if got != "(arr).ptr[i]" {
		t.Errorf("UncheckedIndex = %q, want \"(arr).ptr[i]\"", got)
	}
}

func TestUncheckedIndex_ComplexExpr(t *testing.T) {
	got := cgen.UncheckedIndex("(myslice)", "(idx + 1)")
	if got != "((myslice)).ptr[(idx + 1)]" {
		t.Errorf("UncheckedIndex complex = %q", got)
	}
}

// --- EmitUnsafeBlock Tests ---

func TestEmitUnsafeBlock_SetsUnsafeFlag(t *testing.T) {
	table := types.NewTypeTable()
	intern := ast.NewInternPool(64)
	queue := cgen.NewTypeDeclQueue()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	eg := cgen.NewExprGen(table, intern, nil, nil, queue)

	if eg.Unsafe {
		t.Fatal("ExprGen should not be unsafe initially")
	}

	var innerUnsafe bool
	cgen.EmitUnsafeBlock(w, eg, func() {
		innerUnsafe = eg.Unsafe
	})

	if !innerUnsafe {
		t.Error("ExprGen should be unsafe inside the block")
	}
	if eg.Unsafe {
		t.Error("ExprGen should be restored to safe after the block")
	}
}

func TestEmitUnsafeBlock_Output(t *testing.T) {
	table := types.NewTypeTable()
	intern := ast.NewInternPool(64)
	queue := cgen.NewTypeDeclQueue()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	eg := cgen.NewExprGen(table, intern, nil, nil, queue)

	cgen.EmitUnsafeBlock(w, eg, func() {
		w.Line("ax_u8* raw = some_ptr;")
		w.Line("raw[0] = 255;")
	})

	out := buf.String()
	if !strings.Contains(out, "/* unsafe */") {
		t.Errorf("missing unsafe comment in:\n%s", out)
	}
	if !strings.Contains(out, "ax_u8* raw = some_ptr;") {
		t.Errorf("missing raw decl in:\n%s", out)
	}
	if !strings.Contains(out, "raw[0] = 255;") {
		t.Errorf("missing assignment in:\n%s", out)
	}
	// Check indentation: body should be indented relative to block braces
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines, got %d:\n%s", len(lines), out)
	}
	if lines[0] != "{ /* unsafe */" {
		t.Errorf("first line = %q, want \"{ /* unsafe */\"", lines[0])
	}
	if lines[len(lines)-1] != "}" {
		t.Errorf("last line = %q, want \"}\"", lines[len(lines)-1])
	}
}

func TestEmitUnsafeBlock_NestedRestoresState(t *testing.T) {
	table := types.NewTypeTable()
	intern := ast.NewInternPool(64)
	queue := cgen.NewTypeDeclQueue()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	eg := cgen.NewExprGen(table, intern, nil, nil, queue)

	// Nest unsafe blocks
	cgen.EmitUnsafeBlock(w, eg, func() {
		if !eg.Unsafe {
			t.Error("outer unsafe block: should be unsafe")
		}
		cgen.EmitUnsafeBlock(w, eg, func() {
			if !eg.Unsafe {
				t.Error("inner unsafe block: should be unsafe")
			}
		})
		if !eg.Unsafe {
			t.Error("after inner block: should still be unsafe")
		}
	})
	if eg.Unsafe {
		t.Error("after all blocks: should be safe")
	}
}

// --- EmitArenaBlock Tests ---

func TestEmitArenaBlock_Output(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitArenaBlock(w, "my_arena", func() {
		w.Line("/* allocations here */")
	})

	out := buf.String()
	if !strings.Contains(out, "/* arena block */") {
		t.Errorf("missing arena block comment in:\n%s", out)
	}
	if !strings.Contains(out, "ax_arena_destroy(my_arena);") {
		t.Errorf("missing arena destroy in:\n%s", out)
	}
	if !strings.Contains(out, "/* allocations here */") {
		t.Errorf("missing body content in:\n%s", out)
	}
}

func TestEmitArenaBlock_DestroyAfterClose(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitArenaBlock(w, "arena", func() {
		w.Line("/* body */")
	})

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")

	// destroy should come after the closing brace
	closeBraceIdx := -1
	destroyIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "}" {
			closeBraceIdx = i
		}
		if strings.Contains(line, "ax_arena_destroy") {
			destroyIdx = i
		}
	}
	if closeBraceIdx < 0 {
		t.Fatalf("missing closing brace in:\n%s", out)
	}
	if destroyIdx < 0 {
		t.Fatalf("missing destroy call in:\n%s", out)
	}
	if destroyIdx <= closeBraceIdx {
		t.Errorf("destroy should come after closing brace: brace at %d, destroy at %d", closeBraceIdx, destroyIdx)
	}
}

// --- EmitArenaCreate Tests ---

func TestEmitArenaCreate(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaCreate(w, "my_arena", 4096)
	out := buf.String()
	if !strings.Contains(out, "AxArena* my_arena = ax_arena_create(4096);") {
		t.Errorf("arena create = %q", out)
	}
}

func TestEmitArenaCreate_CustomCapacity(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaCreate(w, "big_arena", 1048576)
	out := buf.String()
	if !strings.Contains(out, "ax_arena_create(1048576)") {
		t.Errorf("arena create with large capacity = %q", out)
	}
}

// --- EmitArenaAlloc Tests ---

func TestEmitArenaAlloc(t *testing.T) {
	table, intern, queue := helper()
	nodeName := intern.InternString("Node")
	nodeID := table.RegisterStruct(nodeName, nil, nil)

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaAlloc(w, "nodes", nodeID, "arena", table, intern, queue)
	out := buf.String()

	if !strings.Contains(out, "struct ax_Node*") {
		t.Errorf("arena alloc missing type: %q", out)
	}
	if !strings.Contains(out, "ax_arena_alloc(arena, sizeof(struct ax_Node))") {
		t.Errorf("arena alloc missing alloc call: %q", out)
	}
}

func TestEmitArenaAlloc_PrimitiveType(t *testing.T) {
	table, intern, queue := helper()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaAlloc(w, "data", types.TypeI32, "arena", table, intern, queue)
	out := buf.String()

	if !strings.Contains(out, "ax_i32* data = (ax_i32*)ax_arena_alloc(arena, sizeof(ax_i32));") {
		t.Errorf("arena alloc primitive = %q", out)
	}
}

// --- EmitArenaAllocInit Tests ---

func TestEmitArenaAllocInit(t *testing.T) {
	table, intern, queue := helper()
	nodeName := intern.InternString("Node")
	nodeID := table.RegisterStruct(nodeName, nil, nil)

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaAllocInit(w, "n", nodeID, "arena", "((struct ax_Node){.x=1})", table, intern, queue)
	out := buf.String()

	if !strings.Contains(out, "ax_arena_alloc(arena, sizeof(struct ax_Node))") {
		t.Errorf("arena alloc init missing alloc: %q", out)
	}
	if !strings.Contains(out, "*n = ((struct ax_Node){.x=1});") {
		t.Errorf("arena alloc init missing assignment: %q", out)
	}
}

// --- EmitArenaReset Tests ---

func TestEmitArenaReset(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaReset(w, "arena")
	out := buf.String()
	if !strings.Contains(out, "ax_arena_reset(arena);") {
		t.Errorf("arena reset = %q", out)
	}
}

// --- EmitArenaDestroy Tests ---

func TestEmitArenaDestroy(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitArenaDestroy(w, "arena")
	out := buf.String()
	if !strings.Contains(out, "ax_arena_destroy(arena);") {
		t.Errorf("arena destroy = %q", out)
	}
}

// --- ArenaAllocExpr Tests ---

func TestArenaAllocExpr(t *testing.T) {
	table, intern, queue := helper()
	nodeName := intern.InternString("Edge")
	edgeID := table.RegisterStruct(nodeName, nil, nil)

	got := cgen.ArenaAllocExpr(edgeID, "arena", table, intern, queue)
	if got != "(struct ax_Edge*)ax_arena_alloc(arena, sizeof(struct ax_Edge))" {
		t.Errorf("ArenaAllocExpr = %q", got)
	}
}

func TestArenaAllocExpr_Primitive(t *testing.T) {
	table, intern, queue := helper()

	got := cgen.ArenaAllocExpr(types.TypeF64, "my_arena", table, intern, queue)
	if got != "(ax_f64*)ax_arena_alloc(my_arena, sizeof(ax_f64))" {
		t.Errorf("ArenaAllocExpr primitive = %q", got)
	}
}

// --- Combined Unsafe + Arena Tests ---

func TestUnsafeInsideArena(t *testing.T) {
	table := types.NewTypeTable()
	intern := ast.NewInternPool(64)
	queue := cgen.NewTypeDeclQueue()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	eg := cgen.NewExprGen(table, intern, nil, nil, queue)

	cgen.EmitArenaBlock(w, "arena", func() {
		cgen.EmitUnsafeBlock(w, eg, func() {
			w.Line("/* unsafe inside arena */")
		})
	})

	out := buf.String()
	if !strings.Contains(out, "/* arena block */") {
		t.Errorf("missing arena comment in:\n%s", out)
	}
	if !strings.Contains(out, "/* unsafe */") {
		t.Errorf("missing unsafe comment in:\n%s", out)
	}
	if !strings.Contains(out, "/* unsafe inside arena */") {
		t.Errorf("missing body in:\n%s", out)
	}
	if !strings.Contains(out, "ax_arena_destroy(arena);") {
		t.Errorf("missing destroy in:\n%s", out)
	}

	// unsafe flag should be restored
	if eg.Unsafe {
		t.Error("unsafe flag not restored after block")
	}
}

func TestNestedArenaBlocks(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitArenaBlock(w, "outer_arena", func() {
		w.Line("/* outer body */")
		cgen.EmitArenaBlock(w, "inner_arena", func() {
			w.Line("/* inner body */")
		})
	})

	out := buf.String()
	// Both destroys should be present
	if strings.Count(out, "ax_arena_destroy") != 2 {
		t.Errorf("expected 2 arena_destroy calls, got:\n%s", out)
	}
	if !strings.Contains(out, "ax_arena_destroy(inner_arena);") {
		t.Errorf("missing inner destroy in:\n%s", out)
	}
	if !strings.Contains(out, "ax_arena_destroy(outer_arena);") {
		t.Errorf("missing outer destroy in:\n%s", out)
	}
}

// --- UnsafeBlockGen/ArenaBlockGen constructor tests ---

func TestNewUnsafeBlockGen(t *testing.T) {
	table, intern, queue := helper()
	gen := cgen.NewUnsafeBlockGen(table, intern, queue)
	if gen == nil {
		t.Error("NewUnsafeBlockGen returned nil")
	}
}

func TestNewArenaBlockGen(t *testing.T) {
	table, intern, queue := helper()
	gen := cgen.NewArenaBlockGen(table, intern, queue)
	if gen == nil {
		t.Error("NewArenaBlockGen returned nil")
	}
}
