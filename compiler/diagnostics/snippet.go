package diagnostics

import (
	"bytes"
	"strings"
)

// ExtractLine returns the content of line N (1-indexed) from src.
// Returns empty string if line is out of bounds.
func ExtractLine(src []byte, line uint32) string {
	if line == 0 {
		return ""
	}
	lines := bytes.Split(src, []byte{'\n'})
	idx := int(line) - 1
	if idx < 0 || idx >= len(lines) {
		return ""
	}
	// Trim trailing \r for CRLF line endings
	return string(bytes.TrimRight(lines[idx], "\r"))
}

// ExtractSnippet returns lines [startLine, endLine] (1-indexed, inclusive) from src.
// Returns nil if range is invalid.
func ExtractSnippet(src []byte, startLine, endLine uint32) []string {
	if startLine == 0 || endLine < startLine {
		return nil
	}
	lines := bytes.Split(src, []byte{'\n'})
	startIdx := int(startLine) - 1
	endIdx := int(endLine) - 1

	if startIdx >= len(lines) {
		return nil
	}
	if endIdx >= len(lines) {
		endIdx = len(lines) - 1
	}

	result := make([]string, 0, endIdx-startIdx+1)
	for i := startIdx; i <= endIdx; i++ {
		result = append(result, string(bytes.TrimRight(lines[i], "\r")))
	}
	return result
}

// Underline returns a string of spaces and carets (^^^) underlining
// columns [startCol, endCol) on the given line content.
// Columns are 1-indexed. Tab characters are expanded to tabWidth spaces.
func Underline(lineContent string, startCol, endCol uint32, tabWidth int) string {
	if tabWidth <= 0 {
		tabWidth = 4
	}
	if startCol == 0 {
		startCol = 1
	}
	if endCol <= startCol {
		endCol = startCol + 1
	}

	// Build the visual representation of the line for alignment
	var prefix strings.Builder
	col := uint32(1)
	for _, ch := range lineContent {
		if col >= startCol {
			break
		}
		if ch == '\t' {
			spaces := tabWidth - (int(col-1) % tabWidth)
			for i := 0; i < spaces; i++ {
				prefix.WriteByte(' ')
			}
			col += uint32(spaces)
		} else {
			prefix.WriteByte(' ')
			col++
		}
	}

	// Calculate caret length
	caretLen := int(endCol - startCol)
	if caretLen <= 0 {
		caretLen = 1
	}

	return prefix.String() + strings.Repeat("^", caretLen)
}
