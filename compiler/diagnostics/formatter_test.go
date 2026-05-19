package diagnostics

import (
	"strings"
	"testing"
)

// noColorOpts creates format options with color disabled for deterministic test output.
func noColorOpts() FormatOptions {
	return FormatOptions{
		UseColor:  false,
		TabWidth:  4,
		MaxLines:  3,
		ShowHints: true,
	}
}

func TestFormatSimpleError(t *testing.T) {
	src := []byte("let x: i32 = \"hello\"\n")
	d := Diagnostic{
		Severity: SeverityError,
		Code:     42,
		Pos:      Pos{Offset: 13, Line: 1, Col: 14},
		Message:  "type mismatch",
	}

	result := FormatDiagnostic(d, src, "main.ax", noColorOpts())

	if !strings.Contains(result, "error[E0042]") {
		t.Errorf("expected error code E0042, got:\n%s", result)
	}
	if !strings.Contains(result, "type mismatch") {
		t.Errorf("expected message 'type mismatch', got:\n%s", result)
	}
	if !strings.Contains(result, "main.ax:1:14") {
		t.Errorf("expected location main.ax:1:14, got:\n%s", result)
	}
	if !strings.Contains(result, "let x: i32 = \"hello\"") {
		t.Errorf("expected source line, got:\n%s", result)
	}
	if !strings.Contains(result, "^") {
		t.Errorf("expected underline caret, got:\n%s", result)
	}
}

func TestFormatWarning(t *testing.T) {
	src := []byte("let x = 42\n")
	d := Diagnostic{
		Severity: SeverityWarning,
		Code:     1001,
		Pos:      Pos{Offset: 4, Line: 1, Col: 5},
		Message:  "unused variable 'x'",
	}

	result := FormatDiagnostic(d, src, "test.ax", noColorOpts())

	if !strings.Contains(result, "warning[E1001]") {
		t.Errorf("expected warning[E1001], got:\n%s", result)
	}
}

func TestFormatNote(t *testing.T) {
	src := []byte("fn foo():\n    return 1\n")
	d := Diagnostic{
		Severity: SeverityNote,
		Code:     2001,
		Pos:      Pos{Offset: 0, Line: 1, Col: 1},
		Message:  "declared here",
	}

	result := FormatDiagnostic(d, src, "test.ax", noColorOpts())

	if !strings.Contains(result, "note[E2001]") {
		t.Errorf("expected note[E2001], got:\n%s", result)
	}
}

func TestFormatWithHint(t *testing.T) {
	src := []byte("let x = 42\n")
	d := Diagnostic{
		Severity: SeverityWarning,
		Code:     1001,
		Pos:      Pos{Offset: 4, Line: 1, Col: 5},
		Message:  "unused variable 'x'",
		Hint:     "consider using '_' prefix",
	}

	result := FormatDiagnostic(d, src, "test.ax", noColorOpts())

	if !strings.Contains(result, "consider using '_' prefix") {
		t.Errorf("expected hint text, got:\n%s", result)
	}
}

func TestFormatHintDisabled(t *testing.T) {
	src := []byte("let x = 42\n")
	d := Diagnostic{
		Severity: SeverityWarning,
		Code:     1001,
		Pos:      Pos{Offset: 4, Line: 1, Col: 5},
		Message:  "unused variable 'x'",
		Hint:     "consider using '_' prefix",
	}

	opts := noColorOpts()
	opts.ShowHints = false
	result := FormatDiagnostic(d, src, "test.ax", opts)

	if strings.Contains(result, "consider using '_' prefix") {
		t.Errorf("hint should be hidden when ShowHints=false, got:\n%s", result)
	}
}

func TestFormatNoColor(t *testing.T) {
	src := []byte("let x = 42\n")
	d := Diagnostic{
		Severity: SeverityError,
		Code:     42,
		Pos:      Pos{Offset: 0, Line: 1, Col: 1},
		Message:  "test error",
	}

	result := FormatDiagnostic(d, src, "test.ax", noColorOpts())

	if strings.Contains(result, "\033[") {
		t.Errorf("expected no ANSI escape codes with UseColor=false, got:\n%s", result)
	}
}

func TestFormatWithColor(t *testing.T) {
	src := []byte("let x = 42\n")
	d := Diagnostic{
		Severity: SeverityError,
		Code:     42,
		Pos:      Pos{Offset: 0, Line: 1, Col: 1},
		Message:  "test error",
	}

	opts := FormatOptions{UseColor: true, TabWidth: 4, ShowHints: true}
	result := FormatDiagnostic(d, src, "test.ax", opts)

	if !strings.Contains(result, "\033[") {
		t.Errorf("expected ANSI escape codes with UseColor=true, got:\n%s", result)
	}
}

func TestFormatICE(t *testing.T) {
	result := FormatICE(
		"compiler/sema/typechecker.go:342",
		"TypeChecker.inferBinaryExpr",
		"BinaryExpr",
		"file.ax",
		Pos{Line: 10, Col: 5},
	)

	if !strings.Contains(result, "internal compiler error") {
		t.Errorf("expected ICE header, got:\n%s", result)
	}
	if !strings.Contains(result, "TypeChecker.inferBinaryExpr") {
		t.Errorf("expected function name, got:\n%s", result)
	}
	if !strings.Contains(result, "BinaryExpr [file.ax:10:5]") {
		t.Errorf("expected node description, got:\n%s", result)
	}
	if !strings.Contains(result, "github.com/axiom-lang/axiom/issues") {
		t.Errorf("expected bug report URL, got:\n%s", result)
	}
}

func TestFormatICENoNode(t *testing.T) {
	result := FormatICE(
		"compiler/lexer/lexer.go:100",
		"Lexer.next",
		"",
		"",
		Pos{},
	)

	if !strings.Contains(result, "internal compiler error") {
		t.Errorf("expected ICE header, got:\n%s", result)
	}
	if strings.Contains(result, "node:") {
		t.Errorf("should not contain node line when nodeDesc is empty, got:\n%s", result)
	}
}

func TestExtractLine(t *testing.T) {
	src := []byte("line 1\nline 2\nline 3\n")

	tests := []struct {
		line uint32
		want string
	}{
		{1, "line 1"},
		{2, "line 2"},
		{3, "line 3"},
		{0, ""},  // out of bounds
		{99, ""}, // out of bounds
	}

	for _, tt := range tests {
		got := ExtractLine(src, tt.line)
		if got != tt.want {
			t.Errorf("ExtractLine(src, %d) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestExtractLineCRLF(t *testing.T) {
	src := []byte("line 1\r\nline 2\r\n")
	got := ExtractLine(src, 1)
	if got != "line 1" {
		t.Errorf("ExtractLine with CRLF = %q, want %q", got, "line 1")
	}
}

func TestExtractLineEmptySource(t *testing.T) {
	got := ExtractLine([]byte{}, 1)
	// Empty source has one empty line
	if got != "" {
		t.Errorf("ExtractLine(empty, 1) = %q, want %q", got, "")
	}
}

func TestExtractSnippet(t *testing.T) {
	src := []byte("line 1\nline 2\nline 3\nline 4\nline 5\n")

	lines := ExtractSnippet(src, 2, 4)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line 2" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "line 2")
	}
	if lines[2] != "line 4" {
		t.Errorf("lines[2] = %q, want %q", lines[2], "line 4")
	}
}

func TestExtractSnippetInvalid(t *testing.T) {
	src := []byte("line 1\n")
	if ExtractSnippet(src, 0, 1) != nil {
		t.Error("expected nil for startLine=0")
	}
	if ExtractSnippet(src, 3, 1) != nil {
		t.Error("expected nil for endLine < startLine")
	}
}

func TestUnderline(t *testing.T) {
	line := "    let x = 42"
	result := Underline(line, 5, 8, 4)

	// Should have 4 spaces (for columns 1-4) then 3 carets
	if !strings.Contains(result, "   ^^^") {
		t.Errorf("expected underline with 3 carets at col 5, got: %q", result)
	}
}

func TestUnderlineWithTab(t *testing.T) {
	line := "\tlet x"
	result := Underline(line, 5, 8, 4)

	// Tab at col 1 expands to 4 spaces, so col 5 = position after tab
	if !strings.Contains(result, "^^^") {
		t.Errorf("expected carets in underline, got: %q", result)
	}
}

func TestDiagnosticSorting(t *testing.T) {
	src := []byte("line 1\nline 2\nline 3\n")
	diags := []Diagnostic{
		{Severity: SeverityError, Code: 1, Pos: Pos{Line: 3, Col: 1}, Message: "error at line 3"},
		{Severity: SeverityError, Code: 2, Pos: Pos{Line: 1, Col: 1}, Message: "error at line 1"},
		{Severity: SeverityError, Code: 3, Pos: Pos{Line: 2, Col: 1}, Message: "error at line 2"},
	}

	result := FormatDiagnostics(diags, src, "test.ax", noColorOpts())

	// Find positions of each error in the output
	pos1 := strings.Index(result, "error at line 1")
	pos2 := strings.Index(result, "error at line 2")
	pos3 := strings.Index(result, "error at line 3")

	if pos1 >= pos2 || pos2 >= pos3 {
		t.Errorf("diagnostics not sorted by line, positions: %d, %d, %d\nresult:\n%s",
			pos1, pos2, pos3, result)
	}
}

func TestDiagnosticDedup(t *testing.T) {
	src := []byte("let x = 1\n")
	diags := []Diagnostic{
		{Severity: SeverityError, Code: 1, Pos: Pos{Line: 1, Col: 5}, Message: "duplicate"},
		{Severity: SeverityError, Code: 1, Pos: Pos{Line: 1, Col: 5}, Message: "duplicate"},
	}

	result := FormatDiagnostics(diags, src, "test.ax", noColorOpts())

	count := strings.Count(result, "duplicate")
	// Should appear in header + source line context, but only once as a diagnostic
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of 'duplicate' message, got %d\nresult:\n%s",
			count, result)
	}
}

func TestFormatDiagnosticsEmpty(t *testing.T) {
	result := FormatDiagnostics(nil, []byte("x"), "f.ax", noColorOpts())
	if result != "" {
		t.Errorf("expected empty string for nil diagnostics, got: %q", result)
	}
}

func TestFormatDiagnosticEmptySource(t *testing.T) {
	d := Diagnostic{
		Severity: SeverityError,
		Code:     1,
		Pos:      Pos{Line: 1, Col: 1},
		Message:  "error in empty file",
	}

	// Should not panic on empty source
	result := FormatDiagnostic(d, []byte{}, "empty.ax", noColorOpts())
	if !strings.Contains(result, "error in empty file") {
		t.Errorf("expected error message in output, got:\n%s", result)
	}
}

func TestSeverityLabel(t *testing.T) {
	if severityLabel(SeverityError) != "error" {
		t.Error("expected 'error' for SeverityError")
	}
	if severityLabel(SeverityWarning) != "warning" {
		t.Error("expected 'warning' for SeverityWarning")
	}
	if severityLabel(SeverityNote) != "note" {
		t.Error("expected 'note' for SeverityNote")
	}
}

func TestDefaultFormatOptions(t *testing.T) {
	opts := DefaultFormatOptions()
	if opts.TabWidth != 4 {
		t.Errorf("TabWidth = %d, want 4", opts.TabWidth)
	}
	if opts.MaxLines != 3 {
		t.Errorf("MaxLines = %d, want 3", opts.MaxLines)
	}
	if !opts.ShowHints {
		t.Error("ShowHints should be true by default")
	}
}
