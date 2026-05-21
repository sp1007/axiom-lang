package cgen_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestEmitHeapDecl(t *testing.T) {
	table, intern, queue := helper()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitHeapDecl(w, "my_foo", types.TypeI32, "42", table, intern, queue)
	output := buf.String()

	// Should emit ax_alloc
	if !strings.Contains(output, "ax_alloc(sizeof(ax_i32))") {
		t.Errorf("missing ax_alloc:\n%s", output)
	}
	// Should init the raw pointer
	if !strings.Contains(output, "*_ax_raw_my_foo = 42;") {
		t.Errorf("missing init:\n%s", output)
	}
	// Should create AxRef
	if !strings.Contains(output, "AxRef my_foo = ax_make_ref(_ax_raw_my_foo);") {
		t.Errorf("missing ax_make_ref:\n%s", output)
	}
}

func TestEmitHeapDecl_NoInit(t *testing.T) {
	table, intern, queue := helper()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitHeapDecl(w, "x", types.TypeF64, "", table, intern, queue)
	output := buf.String()

	// Should NOT have init line
	if strings.Contains(output, "*_ax_raw_x =") {
		t.Errorf("should not have init line without init expr:\n%s", output)
	}
	// Should still have alloc and make_ref
	if !strings.Contains(output, "ax_alloc") {
		t.Errorf("missing ax_alloc:\n%s", output)
	}
	if !strings.Contains(output, "ax_make_ref") {
		t.Errorf("missing ax_make_ref:\n%s", output)
	}
}

func TestEmitStackDecl(t *testing.T) {
	table, intern, queue := helper()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitStackDecl(w, "x", types.TypeI32, "42", table, intern, queue)
	output := buf.String()

	if !strings.Contains(output, "ax_i32 x = 42;") {
		t.Errorf("stack decl = %q", output)
	}
	// Must NOT contain any ax_alloc or AxRef
	if strings.Contains(output, "ax_alloc") {
		t.Errorf("stack decl should not contain ax_alloc:\n%s", output)
	}
	if strings.Contains(output, "AxRef") {
		t.Errorf("stack decl should not contain AxRef:\n%s", output)
	}
}

func TestEmitStackDecl_NoInit(t *testing.T) {
	table, intern, queue := helper()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitStackDecl(w, "x", types.TypeI32, "", table, intern, queue)
	output := buf.String()

	if !strings.Contains(output, "ax_i32 x = {0};") {
		t.Errorf("stack decl no init = %q", output)
	}
}

func TestEmitArenaDecl(t *testing.T) {
	table, intern, queue := helper()

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitArenaDecl(w, "val", types.TypeF32, "3.14f", "_arena", table, intern, queue)
	output := buf.String()

	if !strings.Contains(output, "ax_arena_alloc(_arena, sizeof(ax_f32))") {
		t.Errorf("missing arena alloc:\n%s", output)
	}
	if !strings.Contains(output, "*val = 3.14f;") {
		t.Errorf("missing init:\n%s", output)
	}
	// Must NOT contain AxRef or ax_make_ref
	if strings.Contains(output, "AxRef") {
		t.Errorf("arena decl should not contain AxRef:\n%s", output)
	}
}

func TestHeapFieldAccess(t *testing.T) {
	table, intern, queue := helper()

	fooName := intern.InternString("Foo")
	fooID := table.RegisterStruct(fooName, nil, nil)

	got := cgen.HeapFieldAccess("my_foo", "x", fooID, table, intern, queue)
	if !strings.Contains(got, "ax_deref(my_foo)") {
		t.Errorf("heap field access should contain ax_deref: %q", got)
	}
	if !strings.Contains(got, "->x") {
		t.Errorf("heap field access should use -> operator: %q", got)
	}
}

func TestUnsafeHeapFieldAccess(t *testing.T) {
	table, intern, queue := helper()

	fooName := intern.InternString("Foo")
	fooID := table.RegisterStruct(fooName, nil, nil)

	got := cgen.UnsafeHeapFieldAccess("my_foo", "x", fooID, table, intern, queue)
	// Should NOT contain ax_deref
	if strings.Contains(got, "ax_deref") {
		t.Errorf("unsafe field access should NOT contain ax_deref: %q", got)
	}
	// Should use .ptr direct access
	if !strings.Contains(got, "my_foo.ptr") {
		t.Errorf("unsafe field access should use .ptr: %q", got)
	}
	if !strings.Contains(got, "->x") {
		t.Errorf("unsafe field access should use -> operator: %q", got)
	}
}

func TestArenaFieldAccess(t *testing.T) {
	got := cgen.ArenaFieldAccess("val", "x")
	if got != "val->x" {
		t.Errorf("arena field access = %q, want \"val->x\"", got)
	}
}

func TestStackFieldAccess(t *testing.T) {
	got := cgen.StackFieldAccess("val", "x")
	if got != "val.x" {
		t.Errorf("stack field access = %q, want \"val.x\"", got)
	}
}

func TestHeapDerefExpr_Safe(t *testing.T) {
	table, intern, queue := helper()

	got := cgen.HeapDerefExpr("my_ref", types.TypeI32, false, table, intern, queue)
	if !strings.Contains(got, "ax_deref(my_ref)") {
		t.Errorf("safe heap deref should contain ax_deref: %q", got)
	}
}

func TestHeapDerefExpr_Unsafe(t *testing.T) {
	table, intern, queue := helper()

	got := cgen.HeapDerefExpr("my_ref", types.TypeI32, true, table, intern, queue)
	if strings.Contains(got, "ax_deref") {
		t.Errorf("unsafe heap deref should NOT contain ax_deref: %q", got)
	}
	if !strings.Contains(got, "(my_ref).ptr") {
		t.Errorf("unsafe heap deref should use .ptr: %q", got)
	}
}

func TestFieldAccessForMode_Stack(t *testing.T) {
	table, intern, queue := helper()
	got := cgen.FieldAccessForMode("v", "x", types.TypeI32, cgen.AllocStack, false, table, intern, queue)
	if got != "v.x" {
		t.Errorf("stack mode = %q, want \"v.x\"", got)
	}
}

func TestFieldAccessForMode_Heap(t *testing.T) {
	table, intern, queue := helper()
	got := cgen.FieldAccessForMode("v", "x", types.TypeI32, cgen.AllocHeap, false, table, intern, queue)
	if !strings.Contains(got, "ax_deref(v)") {
		t.Errorf("heap mode should contain ax_deref: %q", got)
	}
}

func TestFieldAccessForMode_HeapUnsafe(t *testing.T) {
	table, intern, queue := helper()
	got := cgen.FieldAccessForMode("v", "x", types.TypeI32, cgen.AllocHeap, true, table, intern, queue)
	if strings.Contains(got, "ax_deref") {
		t.Errorf("heap unsafe mode should NOT contain ax_deref: %q", got)
	}
}

func TestFieldAccessForMode_Arena(t *testing.T) {
	table, intern, queue := helper()
	got := cgen.FieldAccessForMode("v", "x", types.TypeI32, cgen.AllocArena, false, table, intern, queue)
	if got != "v->x" {
		t.Errorf("arena mode = %q, want \"v->x\"", got)
	}
}

func TestAllocContext_DefaultIsStack(t *testing.T) {
	ac := cgen.NewAllocContext()
	if ac.GetMode(42) != cgen.AllocStack {
		t.Error("default alloc mode should be AllocStack")
	}
}

func TestAllocContext_SetGet(t *testing.T) {
	ac := cgen.NewAllocContext()
	ac.SetMode(1, cgen.AllocHeap)
	ac.SetMode(2, cgen.AllocArena)
	ac.SetMode(3, cgen.AllocStack)

	if ac.GetMode(1) != cgen.AllocHeap {
		t.Error("expected AllocHeap for 1")
	}
	if ac.GetMode(2) != cgen.AllocArena {
		t.Error("expected AllocArena for 2")
	}
	if ac.GetMode(3) != cgen.AllocStack {
		t.Error("expected AllocStack for 3")
	}
}
