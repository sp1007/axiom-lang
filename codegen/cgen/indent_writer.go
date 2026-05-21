package cgen

import (
	"fmt"
	"io"
	"strings"
)

// IndentWriter provides indentation-aware writing to an io.Writer.
// It tracks the current indentation level and prefixes each line
// with the appropriate number of spaces.
type IndentWriter struct {
	w      io.Writer
	indent int
}

// NewIndentWriter creates a new IndentWriter wrapping w.
func NewIndentWriter(w io.Writer) *IndentWriter {
	return &IndentWriter{w: w}
}

// Indent increases the indentation level by one.
func (iw *IndentWriter) Indent() { iw.indent++ }

// Dedent decreases the indentation level by one.
func (iw *IndentWriter) Dedent() {
	if iw.indent > 0 {
		iw.indent--
	}
}

// Line writes a single indented line followed by a newline.
func (iw *IndentWriter) Line(s string) {
	fmt.Fprintf(iw.w, "%s%s\n", strings.Repeat("    ", iw.indent), s)
}

// Linef writes a formatted indented line followed by a newline.
func (iw *IndentWriter) Linef(format string, args ...interface{}) {
	iw.Line(fmt.Sprintf(format, args...))
}

// Raw writes a string directly without indentation or newline.
func (iw *IndentWriter) Raw(s string) {
	fmt.Fprint(iw.w, s)
}

// BlankLine writes an empty line.
func (iw *IndentWriter) BlankLine() {
	fmt.Fprintln(iw.w)
}

// Level returns the current indentation level.
func (iw *IndentWriter) Level() int {
	return iw.indent
}
