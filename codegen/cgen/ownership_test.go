package cgen_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestEmitMovePoison_ValueType(t *testing.T) {
	table, intern, _ := helper()
	_ = intern
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	cgen.EmitMovePoison(w, "my_var", types.TypeI32, table)

	output := buf.String()
	if !strings.Contains(output, "#if AX_DEBUG") {
		t.Errorf("missing #if AX_DEBUG guard:\n%s", output)
	}
	if !strings.Contains(output, "memset(&my_var, 0, sizeof(my_var))") {
		t.Errorf("missing memset for value type:\n%s", output)
	}
	if !strings.Contains(output, "#endif") {
		t.Errorf("missing #endif:\n%s", output)
	}
}

func TestEmitParamDecl_LentParam(t *testing.T) {
	table, intern, queue := helper()

	result := cgen.EmitParamDecl("s", types.TypeI32, ast.FlagIsLent, table, intern, queue)
	if result != "const ax_i32* s" {
		t.Errorf("lent param = %q, want \"const ax_i32* s\"", result)
	}
}

func TestEmitParamDecl_MutLentParam(t *testing.T) {
	table, intern, queue := helper()

	result := cgen.EmitParamDecl("s", types.TypeI32, ast.FlagIsLent|ast.FlagIsMut, table, intern, queue)
	if result != "ax_i32* s" {
		t.Errorf("mut lent param = %q, want \"ax_i32* s\"", result)
	}
}

func TestEmitParamDecl_SinkParam(t *testing.T) {
	table, intern, queue := helper()

	result := cgen.EmitParamDecl("x", types.TypeI32, ast.FlagIsSink, table, intern, queue)
	if result != "ax_i32 x" {
		t.Errorf("sink param = %q, want \"ax_i32 x\"", result)
	}
}

func TestEmitParamDecl_ValueParam(t *testing.T) {
	table, intern, queue := helper()

	result := cgen.EmitParamDecl("x", types.TypeI32, 0, table, intern, queue)
	if result != "ax_i32 x" {
		t.Errorf("value param = %q, want \"ax_i32 x\"", result)
	}
}

func TestAdaptArgForParam_Lent(t *testing.T) {
	got := cgen.AdaptArgForParam("my_foo", cgen.ModeRef)
	if got != "&(my_foo)" {
		t.Errorf("adapt for lent = %q, want \"&(my_foo)\"", got)
	}
}

func TestAdaptArgForParam_MutRef(t *testing.T) {
	got := cgen.AdaptArgForParam("my_foo", cgen.ModeMutRef)
	if got != "&(my_foo)" {
		t.Errorf("adapt for mut ref = %q, want \"&(my_foo)\"", got)
	}
}

func TestAdaptArgForParam_Value(t *testing.T) {
	got := cgen.AdaptArgForParam("my_foo", cgen.ModeValue)
	if got != "my_foo" {
		t.Errorf("adapt for value = %q, want \"my_foo\"", got)
	}
}

func TestFieldAccessOp_Value(t *testing.T) {
	if cgen.FieldAccessOp(cgen.ModeValue) != "." {
		t.Error("value mode should use '.'")
	}
}

func TestFieldAccessOp_Ref(t *testing.T) {
	if cgen.FieldAccessOp(cgen.ModeRef) != "->" {
		t.Error("ref mode should use '->'")
	}
}

func TestFieldAccessOp_MutRef(t *testing.T) {
	if cgen.FieldAccessOp(cgen.ModeMutRef) != "->" {
		t.Error("mut ref mode should use '->'")
	}
}

func TestParamModeFromFlags_Lent(t *testing.T) {
	mode := cgen.ParamModeFromFlags(ast.FlagIsLent)
	if mode != cgen.ModeRef {
		t.Errorf("lent flag = mode %d, want ModeRef", mode)
	}
}

func TestParamModeFromFlags_MutLent(t *testing.T) {
	mode := cgen.ParamModeFromFlags(ast.FlagIsLent | ast.FlagIsMut)
	if mode != cgen.ModeMutRef {
		t.Errorf("mut lent flags = mode %d, want ModeMutRef", mode)
	}
}

func TestParamModeFromFlags_Sink(t *testing.T) {
	mode := cgen.ParamModeFromFlags(ast.FlagIsSink)
	if mode != cgen.ModeSink {
		t.Errorf("sink flag = mode %d, want ModeSink", mode)
	}
}

func TestParamModeFromFlags_Value(t *testing.T) {
	mode := cgen.ParamModeFromFlags(0)
	if mode != cgen.ModeValue {
		t.Errorf("no flags = mode %d, want ModeValue", mode)
	}
}

func TestOwnershipContext_DefaultMode(t *testing.T) {
	oc := cgen.NewOwnershipContext()
	if oc.GetMode(42) != cgen.ModeValue {
		t.Error("default mode should be ModeValue")
	}
}

func TestOwnershipContext_SetGet(t *testing.T) {
	oc := cgen.NewOwnershipContext()
	oc.SetMode(10, cgen.ModeRef)
	oc.SetMode(20, cgen.ModeMutRef)

	if oc.GetMode(10) != cgen.ModeRef {
		t.Error("expected ModeRef for nameID 10")
	}
	if oc.GetMode(20) != cgen.ModeMutRef {
		t.Error("expected ModeMutRef for nameID 20")
	}
}
