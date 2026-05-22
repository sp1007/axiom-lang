package cgen

import (
	"testing"
)

func TestLookupBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		isBuiltin bool
		cName    string
		kind     BuiltinKind
	}{
		{"println", true, "ax_println", BuiltinTyped},
		{"print", true, "ax_print", BuiltinTyped},
		{"assert", true, "ax_assert_axiom", BuiltinDirect},
		{"str_len", true, "ax_str_len", BuiltinDirect},
		{"str_contains", true, "ax_str_contains", BuiltinDirect},
		{"gcd", true, "ax_gcd", BuiltinDirect},
		{"lcm", true, "ax_lcm", BuiltinDirect},
		{"sqrt", true, "sqrt", BuiltinDirect},
		{"exit", true, "exit", BuiltinDirect},
		{"panic", true, "ax_panic", BuiltinDirect},
		{"vec_new", true, "ax_vec_new", BuiltinDirect},
		{"arena_new", true, "ax_arena_new", BuiltinDirect},
		// Not builtins
		{"my_func", false, "", BuiltinNone},
		{"foo", false, "", BuiltinNone},
		{"bar_baz", false, "", BuiltinNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := LookupBuiltin(tt.name)
			if ok != tt.isBuiltin {
				t.Errorf("LookupBuiltin(%q) returned ok=%v, want %v", tt.name, ok, tt.isBuiltin)
			}
			if ok {
				if info.CName != tt.cName {
					t.Errorf("LookupBuiltin(%q).CName = %q, want %q", tt.name, info.CName, tt.cName)
				}
				if info.Kind != tt.kind {
					t.Errorf("LookupBuiltin(%q).Kind = %d, want %d", tt.name, info.Kind, tt.kind)
				}
			}
		})
	}
}

func TestEmitBuiltinCall_Direct(t *testing.T) {
	tests := []struct {
		funcName string
		args     []string
		expected string
	}{
		{"gcd", []string{"12", "8"}, "ax_gcd(12, 8)"},
		{"lcm", []string{"a", "b"}, "ax_lcm(a, b)"},
		{"sqrt", []string{"x"}, "sqrt(x)"},
		{"exit", []string{"0"}, "exit(0)"},
		{"panic", []string{`"msg"`}, `ax_panic((const char*)("msg").ptr)`},
		{"str_len", []string{"s"}, "ax_str_len(s)"},
		{"vec_push", []string{"&v", "&elem"}, "ax_vec_push(&v, &elem)"},
		{"arena_new", []string{"4096"}, "ax_arena_new(4096)"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			result := EmitBuiltinCall(tt.funcName, tt.args)
			if result != tt.expected {
				t.Errorf("EmitBuiltinCall(%q, %v) = %q, want %q",
					tt.funcName, tt.args, result, tt.expected)
			}
		})
	}
}

func TestEmitBuiltinCall_Typed(t *testing.T) {
	// Typed builtins append _str suffix by default
	tests := []struct {
		funcName string
		args     []string
		expected string
	}{
		{"println", []string{`AX_STR("hello")`}, `ax_println_str(AX_STR("hello"))`},
		{"print", []string{"s"}, "ax_print_str(s)"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			result := EmitBuiltinCall(tt.funcName, tt.args)
			if result != tt.expected {
				t.Errorf("EmitBuiltinCall(%q, %v) = %q, want %q",
					tt.funcName, tt.args, result, tt.expected)
			}
		})
	}
}

func TestEmitBuiltinCall_NotBuiltin(t *testing.T) {
	result := EmitBuiltinCall("my_custom_func", []string{"a", "b"})
	if result != "" {
		t.Errorf("Expected empty string for non-builtin, got %q", result)
	}
}

func TestIsBuiltin(t *testing.T) {
	if !IsBuiltin("println") {
		t.Error("println should be builtin")
	}
	if !IsBuiltin("gcd") {
		t.Error("gcd should be builtin")
	}
	if IsBuiltin("my_func") {
		t.Error("my_func should not be builtin")
	}
}
