package diagnostics

import "testing"

func TestDiagnosticZeroValue(t *testing.T) {
	var d Diagnostic
	if d.Severity != SeverityError {
		t.Fatalf("expected zero value severity=SeverityError, got %d", d.Severity)
	}
}

func TestDiagnosticMessage(t *testing.T) {
	d := Diagnostic{
		Severity: SeverityError,
		Code:     1001,
		Pos:      Pos{Offset: 42, Line: 3, Col: 7},
		Message:  "undefined variable",
	}
	if d.Error() != "undefined variable" {
		t.Fatalf("Error() = %q, want %q", d.Error(), "undefined variable")
	}
}

func TestDiagnosticSeverityConstants(t *testing.T) {
	// Verify iota ordering is correct
	if SeverityError != 0 {
		t.Fatalf("SeverityError = %d, want 0", SeverityError)
	}
	if SeverityWarning != 1 {
		t.Fatalf("SeverityWarning = %d, want 1", SeverityWarning)
	}
	if SeverityNote != 2 {
		t.Fatalf("SeverityNote = %d, want 2", SeverityNote)
	}
}

func TestDiagnosticHint(t *testing.T) {
	d := Diagnostic{
		Severity: SeverityWarning,
		Code:     2001,
		Pos:      Pos{Offset: 10, Line: 1, Col: 10},
		Message:  "unused variable 'x'",
		Hint:     "consider using '_' prefix for intentionally unused variables",
	}
	if d.Hint == "" {
		t.Fatal("expected non-empty hint")
	}
	if d.Error() != "unused variable 'x'" {
		t.Fatalf("Error() = %q, want %q", d.Error(), "unused variable 'x'")
	}
}

func TestPosFields(t *testing.T) {
	p := Pos{Offset: 100, Line: 5, Col: 12}
	if p.Offset != 100 {
		t.Fatalf("Offset = %d, want 100", p.Offset)
	}
	if p.Line != 5 {
		t.Fatalf("Line = %d, want 5", p.Line)
	}
	if p.Col != 12 {
		t.Fatalf("Col = %d, want 12", p.Col)
	}
}
