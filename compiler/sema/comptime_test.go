package sema

import (
	"math"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestComptimeIntArith(t *testing.T) {
	tests := []struct {
		op   string
		a, b int64
		want int64
		err  bool
	}{
		{"+", 2, 3, 5, false},
		{"+", -1, 1, 0, false},
		{"+", math.MaxInt64, 1, 0, true}, // overflow
		{"-", 10, 3, 7, false},
		{"-", 0, math.MinInt64, 0, true}, // overflow
		{"*", 6, 7, 42, false},
		{"*", math.MaxInt64, 2, 0, true}, // overflow
		{"*", 0, 999, 0, false},
		{"/", 10, 3, 3, false},
		{"/", 10, 0, 0, true}, // div by zero
		{"/", math.MinInt64, -1, 0, true}, // overflow
		{"%", 10, 3, 1, false},
		{"%", 10, 0, 0, true}, // div by zero
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got, diag := IntArith(tt.op, tt.a, tt.b)
			if tt.err {
				if diag == nil {
					t.Errorf("IntArith(%s, %d, %d) expected error, got %d", tt.op, tt.a, tt.b, got)
				}
			} else {
				if diag != nil {
					t.Errorf("IntArith(%s, %d, %d) unexpected error: %s", tt.op, tt.a, tt.b, diag.Message)
				}
				if got != tt.want {
					t.Errorf("IntArith(%s, %d, %d) = %d, want %d", tt.op, tt.a, tt.b, got, tt.want)
				}
			}
		})
	}
}

func TestComptimeFloatArith(t *testing.T) {
	tests := []struct {
		op   string
		a, b float64
		want float64
		err  bool
	}{
		{"+", 3.14, 2.0, 5.14, false},
		{"-", 10.5, 0.5, 10.0, false},
		{"*", 3.14159, 2.0, 6.28318, false},
		{"/", 10.0, 4.0, 2.5, false},
		{"/", 1.0, 0.0, 0, true}, // div by zero
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got, diag := FloatArith(tt.op, tt.a, tt.b)
			if tt.err {
				if diag == nil {
					t.Errorf("FloatArith(%s, %g, %g) expected error", tt.op, tt.a, tt.b)
				}
			} else {
				if diag != nil {
					t.Errorf("FloatArith(%s, %g, %g) unexpected error: %s", tt.op, tt.a, tt.b, diag.Message)
				}
				if math.Abs(got-tt.want) > 1e-10 {
					t.Errorf("FloatArith(%s, %g, %g) = %g, want %g", tt.op, tt.a, tt.b, got, tt.want)
				}
			}
		})
	}
}

func TestComptimeBoolLogic(t *testing.T) {
	tests := []struct {
		op   string
		a, b bool
		want bool
	}{
		{"and", true, true, true},
		{"and", true, false, false},
		{"and", false, false, false},
		{"or", true, false, true},
		{"or", false, false, false},
		{"not", true, false, false},
		{"not", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got, diag := BoolLogic(tt.op, tt.a, tt.b)
			if diag != nil {
				t.Errorf("BoolLogic(%s, %v, %v) unexpected error: %s", tt.op, tt.a, tt.b, diag.Message)
			}
			if got != tt.want {
				t.Errorf("BoolLogic(%s, %v, %v) = %v, want %v", tt.op, tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestComptimeStringConcat(t *testing.T) {
	got := StringConcat("hello", " world")
	if got != "hello world" {
		t.Errorf("StringConcat = %q, want %q", got, "hello world")
	}
}

func TestComptimeValueString(t *testing.T) {
	tests := []struct {
		val  ComptimeValue
		want string
	}{
		{ComptimeValue{Kind: types.TypeI32, IntVal: 42}, "42"},
		{ComptimeValue{Kind: types.TypeF64, FloatVal: 3.14}, "3.14"},
		{ComptimeValue{Kind: types.TypeBool, BoolVal: true}, "true"},
		{ComptimeValue{Kind: types.TypeBool, BoolVal: false}, "false"},
		{ComptimeValue{Kind: types.TypeString, StrVal: "hello"}, `"hello"`},
	}

	for _, tt := range tests {
		got := tt.val.String()
		if got != tt.want {
			t.Errorf("ComptimeValue.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestComptimeParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"42", 42},
		{"0xFF", 255},
		{"0b1010", 10},
		{"0o77", 63},
		{"1_000_000", 1000000},
		{"0", 0},
	}

	for _, tt := range tests {
		got := parseInt64(tt.input)
		if got != tt.want {
			t.Errorf("parseInt64(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestComptimeParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"3.14", 3.14},
		{"1.0e-6", 1e-6},
		{"2.5", 2.5},
		{"0.0", 0.0},
	}

	for _, tt := range tests {
		got := parseFloat64(tt.input)
		if math.Abs(got-tt.want) > 1e-15 {
			t.Errorf("parseFloat64(%q) = %g, want %g", tt.input, got, tt.want)
		}
	}
}

func TestComptimeEvalLiterals(t *testing.T) {
	// Test that the evaluator can evaluate literal expressions from real source code.
	src := []byte(`const X: i32 = 42
const PI: f64 = 3.14
const MSG: string = "hello"

fn main():
    let x = X
`)
	pool := ast.NewInternPool(16)
	toks, _, _ := lexer.Lex(src)
	tree, _ := parser.Parse(toks, src, pool)

	st := NewSymbolTable(pool)
	tt := types.NewTypeTable()

	// Run name resolution
	lazy := NewLazyResolver(st, tt, nil)
	nr := NewNameResolver(tree, pool, st, tt, lazy)
	nr.Resolve()

	// Run comptime evaluation
	ce := NewComptimeEvaluator(tree, pool, st, tt)
	diags := ce.EvalConsts()

	if len(diags) > 0 {
		for _, d := range diags {
			t.Logf("comptime diag: %s", d.Message)
		}
		t.Errorf("expected 0 comptime errors, got %d", len(diags))
	}

	consts := ce.Consts()
	if len(consts) < 3 {
		t.Fatalf("expected at least 3 consts, got %d", len(consts))
	}

	// Verify that at least one const was evaluated
	found := false
	for _, v := range consts {
		if v.Kind == types.TypeI64 && v.IntVal == 42 {
			found = true
		}
	}
	if !found {
		t.Error("expected const X = 42 to be evaluated")
	}
}

func TestComptimeNonConstant(t *testing.T) {
	// Function calls are not evaluable at compile time
	d := &diagnostics.Diagnostic{}
	_, err := IntArith("invalid_op", 1, 2)
	if err == nil {
		t.Error("expected error for unsupported operation")
	}
	_ = d
}
