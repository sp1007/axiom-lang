// Package diagnostics provides the shared diagnostic types used by all
// compiler passes. Compiler passes MUST NOT panic; they return []Diagnostic
// instead. Diagnostics are product features — they must be deterministic,
// human-readable, source-located, and actionable.
package diagnostics

// Severity classifies how serious a diagnostic is.
type Severity uint8

const (
	SeverityError   Severity = iota
	SeverityWarning
	SeverityNote
)

// Pos identifies a byte offset in the source file.
type Pos struct {
	Offset uint32
	Line   uint32
	Col    uint32
}

// Diagnostic is a compiler message attached to a source location.
// Compiler passes MUST NOT panic; they return []Diagnostic instead.
type Diagnostic struct {
	Severity Severity
	Code     uint32
	Pos      Pos
	Message  string
	Hint     string // optional actionable hint
}

// Error implements the error interface, returning the diagnostic message.
func (d *Diagnostic) Error() string { return d.Message }
