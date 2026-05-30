package driver

import (
	"testing"
)

func TestCompileSimpleProgram(t *testing.T) {
	src := []byte(`fn main():
    let x: i32 = 42
    return
`)

	result := Compile(src, "test.ax", nil)

	if result.Phase != PhaseFull {
		// Log diagnostics for debugging
		for _, d := range result.Diags {
			t.Logf("diag: %s (severity=%d, code=%d)", d.Message, d.Severity, d.Code)
		}
		t.Logf("pipeline stopped at phase: %s", result.Phase)
	}

	// Verify intermediate products were created
	if result.Tokens == nil {
		t.Fatal("expected tokens to be non-nil")
	}
	if result.Tree == nil {
		t.Fatal("expected tree to be non-nil")
	}
	if result.Intern == nil {
		t.Fatal("expected intern pool to be non-nil")
	}
	if result.Symbols == nil {
		t.Fatal("expected symbol table to be non-nil")
	}
	if result.Types == nil {
		t.Fatal("expected type table to be non-nil")
	}
}

func TestCompileLexErrors(t *testing.T) {
	// Unterminated string should produce a lex-level error
	src := []byte(`fn main():
    let x = "unterminated
`)

	result := Compile(src, "test_lex_error.ax", nil)

	if !result.HasErrors() {
		t.Fatal("expected errors from unterminated string")
	}
}

func TestCompileStopAfterParse(t *testing.T) {
	src := []byte(`fn main():
    let x: i32 = 42
    return
`)

	result := Compile(src, "test.ax", &CompileOptions{
		StopAfter:    PhaseParse,
		StopAfterSet: true,
	})

	if result.Phase != PhaseParse {
		t.Fatalf("expected phase PhaseParse, got %s", result.Phase)
	}
	if result.Tree == nil {
		t.Fatal("expected tree to be populated after parse phase")
	}
	// Symbols and Types should NOT be populated since we stopped after parse
	if result.Symbols != nil {
		t.Fatal("expected symbols to be nil when stopped after parse")
	}
}

func TestCompileStopAfterLex(t *testing.T) {
	src := []byte(`fn main():
    return
`)

	result := Compile(src, "test.ax", &CompileOptions{
		StopAfter:    PhaseLex,
		StopAfterSet: true,
	})

	if result.Phase != PhaseLex {
		t.Fatalf("expected phase PhaseLex, got %s", result.Phase)
	}
	if result.Tokens == nil {
		t.Fatal("expected tokens to be populated after lex phase")
	}
	if result.Tree != nil {
		t.Fatal("expected tree to be nil when stopped after lex")
	}
}

func TestCompilePhaseString(t *testing.T) {
	tests := []struct {
		phase CompilePhase
		want  string
	}{
		{PhaseLex, "lex"},
		{PhaseParse, "parse"},
		{PhaseResolve, "resolve"},
		{PhaseInfer, "infer"},
		{PhaseTypeCheck, "typecheck"},
		{PhaseOwnership, "ownership"},
		{PhaseFull, "full"},
		{CompilePhase(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.phase.String()
		if got != tt.want {
			t.Errorf("CompilePhase(%d).String() = %q, want %q", tt.phase, got, tt.want)
		}
	}
}

func TestResolvePositions(t *testing.T) {
	src := []byte(`fn main():
    let x = "unterminated
`)

	result := Compile(src, "test.ax", nil)

	// Before resolving, some diags might only have offset
	result.ResolvePositions()

	// After resolving, line numbers should be filled in
	for _, d := range result.Diags {
		if d.Pos.Offset > 0 && d.Pos.Line == 0 {
			t.Errorf("diagnostic at offset %d still has line=0 after ResolvePositions", d.Pos.Offset)
		}
	}
}
