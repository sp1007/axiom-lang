package codegen_test

import (
	"testing"

	"github.com/axiom-lang/axiom/codegen"
)

func TestMangleBasic(t *testing.T) {
	// math::add(i32, i32) -> i32
	got := codegen.Mangle("math", "add", []uint32{3, 3}, 3)
	expected := "_AX_math_add_ii_i"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}
}

func TestMangleVoid(t *testing.T) {
	// main::main() -> void
	got := codegen.Mangle("main", "main", nil, 0)
	expected := "_AX_main_main_v_v"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}
}

func TestMangleMixedTypes(t *testing.T) {
	// io::write(str, i64) -> bool
	got := codegen.Mangle("io", "write", []uint32{13, 4}, 2)
	expected := "_AX_io_write_tl_o"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}
}

func TestMangleGeneric(t *testing.T) {
	// List[i32]::push(i32) -> void
	got := codegen.MangleGeneric("collections", "push", []uint32{3}, []uint32{3}, 0)
	if got != "_AX_collections_push_Ti_i_v" {
		t.Errorf("got %q", got)
	}
}

func TestMangleMethod(t *testing.T) {
	// Vec::len() -> i32
	got := codegen.MangleMethod("std", "Vec", "len", nil, 3)
	expected := "_AX_std_Vec_len_v_i"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}
}

func TestDemangleBasic(t *testing.T) {
	result, err := codegen.Demangle("_AX_math_add_ii_i")
	if err != nil {
		t.Fatal(err)
	}
	if result.Module != "math" {
		t.Errorf("module = %q", result.Module)
	}
	if result.Name != "add" {
		t.Errorf("name = %q", result.Name)
	}
	if len(result.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(result.Params))
	}
	if result.Ret != 3 {
		t.Errorf("ret = %d, expected 3 (i32)", result.Ret)
	}
}

func TestDemangleRoundtrip(t *testing.T) {
	tests := []struct {
		module string
		name   string
		params []uint32
		ret    uint32
	}{
		{"math", "add", []uint32{3, 3}, 3},
		{"main", "main", nil, 0},
		{"io", "write", []uint32{13, 4}, 2},
		{"core", "noop", nil, 0},
	}

	for _, tt := range tests {
		mangled := codegen.Mangle(tt.module, tt.name, tt.params, tt.ret)
		result, err := codegen.Demangle(mangled)
		if err != nil {
			t.Errorf("Demangle(%q) error: %v", mangled, err)
			continue
		}
		if result.Module != tt.module {
			t.Errorf("module: got %q, want %q", result.Module, tt.module)
		}
		if result.Name != tt.name {
			t.Errorf("name: got %q, want %q", result.Name, tt.name)
		}
	}
}

func TestDemangleInvalid(t *testing.T) {
	_, err := codegen.Demangle("not_mangled")
	if err == nil {
		t.Error("expected error for non-AXIOM symbol")
	}
}

func TestIsMangled(t *testing.T) {
	if !codegen.IsMangled("_AX_math_add_ii_i") {
		t.Error("should recognize mangled name")
	}
	if codegen.IsMangled("printf") {
		t.Error("should not recognize unmangled name")
	}
}

func TestMangleSanitize(t *testing.T) {
	// Module with special chars
	got := codegen.Mangle("my-mod", "fn!", nil, 0)
	if got == "" {
		t.Error("expected non-empty result")
	}
	// Should not contain special chars
	for _, ch := range got {
		if ch == '-' || ch == '!' {
			t.Errorf("mangled name contains special char: %q", got)
		}
	}
}
