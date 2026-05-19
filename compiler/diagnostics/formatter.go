package diagnostics

import (
	"fmt"
	"sort"
	"strings"
)

// Error code base ranges for each subsystem.
const (
	ECodeLexerBase   = 1000 // E1000–E1099: lexer errors
	ECodeParserBase  = 1100 // E1100–E1199: parser errors
	ECodeSemaBase    = 1200 // E1200–E1399: semantic analysis
	ECodeTypeBase    = 1400 // E1400–E1599: type system
	ECodeOwnerBase   = 1600 // E1600–E1799: ownership
	ECodeCodegenBase = 1800 // E1800–E1899: codegen
	ECodeICEBase     = 9000 // E9000+: internal compiler errors
)

// FormatOptions controls how diagnostics are rendered.
type FormatOptions struct {
	UseColor  bool // Enable ANSI color codes
	TabWidth  int  // Tab expansion width (default: 4)
	MaxLines  int  // Max context lines around error (default: 3)
	ShowHints bool // Show hint text if present (default: true)
}

// DefaultFormatOptions returns options for terminal output with auto-detected color support.
func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		UseColor:  ColorSupport(),
		TabWidth:  4,
		MaxLines:  3,
		ShowHints: true,
	}
}

// FormatDiagnostic renders a single diagnostic with source context.
// src is the source file content (for extracting lines).
// filename is the display name of the file.
func FormatDiagnostic(d Diagnostic, src []byte, filename string, opts FormatOptions) string {
	if opts.TabWidth <= 0 {
		opts.TabWidth = 4
	}

	var b strings.Builder

	// Severity label and error code
	sevLabel := severityLabel(d.Severity)
	sevColor := severityColor(d.Severity)

	b.WriteString(colorize(fmt.Sprintf("%s[E%04d]", sevLabel, d.Code), sevColor, opts.UseColor))
	b.WriteString(": ")
	b.WriteString(colorize(d.Message, ColorBold, opts.UseColor))
	b.WriteByte('\n')

	// Location line
	b.WriteString(colorize("  --> ", ColorCyan, opts.UseColor))
	b.WriteString(fmt.Sprintf("%s:%d:%d", filename, d.Pos.Line, d.Pos.Col))
	b.WriteByte('\n')

	// Source snippet with underline
	if len(src) > 0 && d.Pos.Line > 0 {
		lineContent := ExtractLine(src, d.Pos.Line)
		lineNumStr := fmt.Sprintf("%d", d.Pos.Line)
		padding := strings.Repeat(" ", len(lineNumStr))

		// Empty gutter line
		b.WriteString(colorize(fmt.Sprintf(" %s |", padding), ColorDim, opts.UseColor))
		b.WriteByte('\n')

		// Source line
		b.WriteString(colorize(fmt.Sprintf(" %s | ", lineNumStr), ColorDim, opts.UseColor))
		b.WriteString(lineContent)
		b.WriteByte('\n')

		// Underline with hint
		underline := Underline(lineContent, d.Pos.Col, d.Pos.Col+1, opts.TabWidth)
		b.WriteString(colorize(fmt.Sprintf(" %s | ", padding), ColorDim, opts.UseColor))
		hintText := ""
		if opts.ShowHints && d.Hint != "" {
			hintText = " " + d.Hint
		}
		b.WriteString(colorize(underline+hintText, sevColor, opts.UseColor))
		b.WriteByte('\n')

		// Closing gutter
		b.WriteString(colorize(fmt.Sprintf(" %s |", padding), ColorDim, opts.UseColor))
		b.WriteByte('\n')
	}

	return b.String()
}

// FormatDiagnostics renders multiple diagnostics, deduplicating by position
// and sorting by (line, col, severity).
func FormatDiagnostics(diags []Diagnostic, src []byte, filename string, opts FormatOptions) string {
	if len(diags) == 0 {
		return ""
	}

	// Sort by position
	sorted := make([]Diagnostic, len(diags))
	copy(sorted, diags)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Pos.Line != sorted[j].Pos.Line {
			return sorted[i].Pos.Line < sorted[j].Pos.Line
		}
		if sorted[i].Pos.Col != sorted[j].Pos.Col {
			return sorted[i].Pos.Col < sorted[j].Pos.Col
		}
		return sorted[i].Severity < sorted[j].Severity
	})

	// Deduplicate by (position, message)
	deduped := make([]Diagnostic, 0, len(sorted))
	for i, d := range sorted {
		if i > 0 {
			prev := sorted[i-1]
			if d.Pos == prev.Pos && d.Message == prev.Message && d.Severity == prev.Severity {
				continue
			}
		}
		deduped = append(deduped, d)
	}

	var b strings.Builder
	for _, d := range deduped {
		b.WriteString(FormatDiagnostic(d, src, filename, opts))
	}
	return b.String()
}

// FormatICE renders an Internal Compiler Error with context information.
// This format is distinct from normal diagnostics and includes a bug report URL.
func FormatICE(component string, function string, nodeDesc string, filename string, pos Pos) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("axc: internal compiler error at %s\n", component))
	b.WriteString(fmt.Sprintf("  function: %s\n", function))
	if nodeDesc != "" {
		b.WriteString(fmt.Sprintf("  node: %s [%s:%d:%d]\n", nodeDesc, filename, pos.Line, pos.Col))
	}
	b.WriteString("  please report at https://github.com/axiom-lang/axiom/issues\n")
	return b.String()
}

// severityLabel returns the human-readable label for a severity level.
func severityLabel(s Severity) string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityNote:
		return "note"
	default:
		return "unknown"
	}
}

// severityColor returns the ANSI color code for a severity level.
func severityColor(s Severity) string {
	switch s {
	case SeverityError:
		return ColorRed
	case SeverityWarning:
		return ColorYellow
	case SeverityNote:
		return ColorCyan
	default:
		return ColorReset
	}
}
