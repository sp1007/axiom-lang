# p01-t06: Diagnostic Formatter

## Purpose
Implement the diagnostic rendering pipeline that converts raw `Diagnostic` structs into user-facing error messages with source snippets, line/column annotations, colored output, and underline markers. This is the standard error presentation layer used by every compiler pass — from lexer errors through codegen failures.

## Context
AXIOM's compiler produces `[]Diagnostic` from each pass (lexer, parser, type checker, ownership checker). These must be rendered consistently in a format inspired by Rust's `rustc` and Go's `golangci-lint`:

```
error[E0042]: type mismatch
  --> main.ax:42:7
   |
42 |     let x: i32 = "hello"
   |                   ^^^^^^^ expected i32, found string
   |
note: function signature requires i32 return type
  --> main.ax:38:1
   |
38 | fn compute(a: i32) -> i32:
   |                       ^^^ declared here
```

Internal Compiler Errors (ICE) must use a distinct format:
```
axc: internal compiler error at compiler/sema/typechecker.go:342
  function: TypeChecker.inferBinaryExpr
  node: BinaryExpr [file.ax:10:5]
  please report at https://github.com/axiom-lang/axiom/issues
```

## Inputs
- `compiler/diagnostics/diagnostics.go` from p01-t01 — `Diagnostic`, `Pos`, `Severity` types
- Source file bytes (for extracting snippet lines)
- Terminal capabilities (color support detection)

## Outputs
- `compiler/diagnostics/formatter.go` — `FormatDiagnostic()`, `FormatDiagnostics()`, `FormatICE()`
- `compiler/diagnostics/colors.go` — ANSI color helpers with terminal detection
- `compiler/diagnostics/snippet.go` — source line extraction and underline generation
- `compiler/diagnostics/formatter_test.go` — golden tests for formatted output

## Dependencies
- p01-t01: repository-bootstrap — `Diagnostic` struct defined

## Subsystems Affected
- All compiler passes: every pass that emits diagnostics uses this formatter for output
- CLI (`cmd/axc/`): all commands use `FormatDiagnostics()` to print errors
- LSP server (p17-t03): uses the raw `Diagnostic` struct, not the formatted text (but formatter useful for LSP `relatedInformation`)

## Detailed Requirements

### 1. Core Formatter API

```go
package diagnostics

// FormatDiagnostic renders a single diagnostic with source context.
// src is the source file content (for extracting lines).
// filename is the display name of the file.
func FormatDiagnostic(d Diagnostic, src []byte, filename string, opts FormatOptions) string

// FormatDiagnostics renders multiple diagnostics, deduplicating by position.
func FormatDiagnostics(diags []Diagnostic, src []byte, filename string, opts FormatOptions) string

// FormatICE renders an Internal Compiler Error with stack trace context.
func FormatICE(component string, function string, nodeDesc string, filename string, pos Pos) string

type FormatOptions struct {
    UseColor  bool   // Enable ANSI color codes
    TabWidth  int    // Tab expansion width (default: 4)
    MaxLines  int    // Max context lines around error (default: 3)
    ShowHints bool   // Show hint text if present (default: true)
}

// DefaultFormatOptions returns options for terminal output with auto-detected color support.
func DefaultFormatOptions() FormatOptions
```

### 2. Source Snippet Extraction

```go
// ExtractLine returns the content of line N (1-indexed) from src.
func ExtractLine(src []byte, line uint32) string

// ExtractSnippet returns lines [startLine, endLine] from src.
func ExtractSnippet(src []byte, startLine, endLine uint32) []string

// Underline returns a string of spaces and carets (^^^) underlining columns [startCol, endCol).
func Underline(lineContent string, startCol, endCol uint32, tabWidth int) string
```

### 3. Color Support

```go
// ColorSupport detects whether the terminal supports ANSI colors.
// Checks: NO_COLOR env var, TERM=dumb, and os.Stdout.IsTerminal().
func ColorSupport() bool

// Color constants for diagnostic rendering.
const (
    ColorRed     = "\033[31m"    // errors
    ColorYellow  = "\033[33m"    // warnings
    ColorCyan    = "\033[36m"    // notes
    ColorBold    = "\033[1m"     // severity labels, file paths
    ColorReset   = "\033[0m"
    ColorDim     = "\033[2m"     // line numbers
)
```

### 4. Output Format Specification

**Error format:**
```
{severity}[{code}]: {message}
  --> {filename}:{line}:{col}
   |
{line} | {source_line_content}
   |   {underline} {hint}
   |
```

**Multi-line format (when span crosses lines):**
```
error[E0100]: unterminated string literal
  --> main.ax:10:5
   |
10 |     let s = "hello
11 |     world
   |         ^ unterminated string
   |
```

**ICE format:**
```
axc: internal compiler error at {component}:{line}
  function: {function}
  node: {node_description}
  please report at https://github.com/axiom-lang/axiom/issues
```

### 5. Error Code Registry

Define error code ranges for each subsystem:
```go
const (
    ECodeLexerBase   = 1000  // E1000–E1099: lexer errors
    ECodeParserBase  = 1100  // E1100–E1199: parser errors
    ECodeSemaBase    = 1200  // E1200–E1399: semantic analysis
    ECodeTypeBase    = 1400  // E1400–E1599: type system
    ECodeOwnerBase   = 1600  // E1600–E1799: ownership
    ECodeCodegenBase = 1800  // E1800–E1899: codegen
    ECodeICEBase     = 9000  // E9000+: internal compiler errors
)
```

### 6. Determinism

Output must be deterministic: same diagnostics + same source → identical formatted output. No timestamps, no random ordering. Diagnostics are sorted by `(filename, line, col, severity)` before rendering.

### 7. NO_COLOR Standard

If `NO_COLOR` environment variable is set (any value), disable all ANSI codes. This follows the [no-color.org](https://no-color.org/) standard.

## Implementation Steps

1. Create `compiler/diagnostics/colors.go`:
   - Implement `ColorSupport()` checking `NO_COLOR`, `TERM`, and `os.Stdout.Fd()` via `term.IsTerminal()`
   - Define color constants
   - Create `colorize(text, color string, useColor bool) string` helper

2. Create `compiler/diagnostics/snippet.go`:
   - Implement `ExtractLine()` using `bytes.Split(src, []byte{'\n'})` with bounds checking
   - Implement `ExtractSnippet()` for multi-line ranges
   - Implement `Underline()` handling tab expansion

3. Create `compiler/diagnostics/formatter.go`:
   - Implement `FormatDiagnostic()` composing severity label + source snippet + underline
   - Implement `FormatDiagnostics()` with deduplication and sorting
   - Implement `FormatICE()` for internal errors
   - Implement `DefaultFormatOptions()` with auto-detected color

4. Create `compiler/diagnostics/formatter_test.go`:
   - Golden test: single error with source snippet
   - Golden test: warning with hint
   - Golden test: multi-line span
   - Golden test: ICE format
   - Test: NO_COLOR disables ANSI
   - Test: deduplication (same position, same message → only one output)
   - Test: sorting by position

5. Add Go dependency: `golang.org/x/term` for `IsTerminal()` detection.

## Test Plan

### Unit Tests
- `TestFormatSimpleError`: single error at known position → matches golden string
- `TestFormatWarning`: warning severity uses yellow color
- `TestFormatNote`: note severity uses cyan color
- `TestFormatWithHint`: hint text appears on a separate line below the underline
- `TestFormatMultiLine`: span crossing 2 lines renders both lines
- `TestFormatNoColor`: `NO_COLOR=1` → no ANSI escape codes in output
- `TestFormatICE`: ICE format matches template exactly
- `TestExtractLine`: line 1, middle line, last line, out-of-bounds → empty
- `TestUnderline`: tab expansion, UTF-8 multi-byte characters
- `TestDiagnosticSorting`: diagnostics sorted by position
- `TestDiagnosticDedup`: duplicate diagnostics removed

### Golden Tests
- `tests/golden/diagnostics/simple_error.expected` — pre-rendered error output
- `tests/golden/diagnostics/multiline.expected` — multi-line span output
- `tests/golden/diagnostics/ice.expected` — ICE output

### Property Tests
- For any valid `Pos` within source bounds, `FormatDiagnostic` does not panic
- For empty source (`[]byte{}`), `FormatDiagnostic` produces output without crashing
- `FormatDiagnostics(nil, ...)` returns empty string

## Validation Checklist

- [ ] `FormatDiagnostic` produces correct output for all severity levels
- [ ] Source snippets extracted correctly for lines 1, N, and last line
- [ ] Underline markers align with column positions (including tabs)
- [ ] ANSI colors disabled when `NO_COLOR` is set
- [ ] ICE format matches the template from the spec
- [ ] Error codes follow the subsystem range convention
- [ ] Diagnostics sorted deterministically by position
- [ ] Multi-byte UTF-8 characters don't misalign underlines
- [ ] `go test ./compiler/diagnostics/` passes

## Acceptance Criteria

- `FormatDiagnostic` output matches the format shown in plan §5.3
- ICE output matches the exact template from plan §5.3
- All golden tests pass
- `NO_COLOR=1` produces clean output without escape codes

## Definition of Done

- [ ] `compiler/diagnostics/formatter.go` implemented
- [ ] `compiler/diagnostics/snippet.go` implemented
- [ ] `compiler/diagnostics/colors.go` implemented
- [ ] All unit tests pass: `go test ./compiler/diagnostics/ -run TestFormat`
- [ ] Golden tests committed and passing
- [ ] Error code ranges documented in code comments

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| UTF-8 multi-byte characters cause underline misalignment | Use `unicode/utf8.RuneCountInString` for column counting instead of `len()` |
| Terminal width truncation for long lines | Add optional line wrapping in `FormatOptions` (not required for MVP) |
| `golang.org/x/term` dependency adds bloat | It's a small, well-maintained package; acceptable for bootstrap compiler |

## Future Follow-up Tasks

- p02-t04: Lexer error recovery uses `FormatDiagnostics` for error output
- p03-t07: Parser error recovery produces diagnostics rendered by this formatter
- p04-t06: Type checker errors displayed via this formatter
- p08-t09: Build pipeline uses `FormatDiagnostics` for all error display
- p17-t03: LSP server may use raw `Diagnostic` struct but not the formatted text
