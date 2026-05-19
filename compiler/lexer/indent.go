package lexer

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

// ProcessIndentation transforms a raw token stream (with NEWLINE tokens but no
// INDENT/DEDENT) into a stream with explicit block boundaries. It counts leading
// spaces after each NEWLINE, compares against an indent stack, and emits
// INDENT/DEDENT tokens. AXIOM requires exactly 4-space indentation per level.
//
// This function receives the source bytes (for counting spaces), the raw token
// stream from the lexer, and the LineTable for error position reporting.
func ProcessIndentation(src []byte, tokens []Token, lt *LineTable) ([]Token, []diagnostics.Diagnostic) {
	if len(tokens) == 0 {
		return tokens, nil
	}

	out := make([]Token, 0, len(tokens)+32)
	var diags []diagnostics.Diagnostic
	stack := indentStack{levels: []int{0}}

	i := 0
	for i < len(tokens) {
		tok := tokens[i]

		// Non-NEWLINE tokens pass through directly
		if tok.Kind != TokenNewline {
			// If it's EOF, first emit remaining DEDENTs
			if tok.Kind == TokenEOF {
				for stack.top() > 0 {
					out = append(out, syntheticToken(TokenDedent, tok))
					stack.pop()
				}
				out = append(out, tok)
				i++
				continue
			}
			out = append(out, tok)
			i++
			continue
		}

		// We found a NEWLINE. Look ahead to find the indent level of the next
		// non-blank, non-comment-only line.
		nextStart := findNextContentOffset(tokens, i+1, src)
		if nextStart < 0 {
			// No more real content — emit the NEWLINE and let EOF handling
			// take care of DEDENTs
			out = append(out, tok)
			i++
			continue
		}

		nextIndent := countLeadingSpaces(src, nextStart)
		currentIndent := stack.top()

		if nextIndent > currentIndent {
			diff := nextIndent - currentIndent
			if diff != 4 {
				line, col := lt.LineCol(uint32(nextStart))
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Code:     10,
					Pos:      diagnostics.Pos{Offset: uint32(nextStart), Line: line, Col: col},
					Message: formatIndentError(
						"indentation increase is not a multiple of 4 spaces (got %d, from level %d)",
						diff, currentIndent),
				})
			}
			stack.push(nextIndent)
			out = append(out, tok) // keep NEWLINE
			out = append(out, syntheticToken(TokenIndent, tok))
		} else if nextIndent < currentIndent {
			out = append(out, tok) // keep NEWLINE
			for stack.top() > nextIndent {
				out = append(out, syntheticToken(TokenDedent, tok))
				stack.pop()
			}
			if stack.top() != nextIndent {
				line, col := lt.LineCol(uint32(nextStart))
				diags = append(diags, diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Code:     11,
					Pos:      diagnostics.Pos{Offset: uint32(nextStart), Line: line, Col: col},
					Message:  "unindent does not match any enclosing indentation level",
				})
			}
		} else {
			// Same level — just keep the NEWLINE
			out = append(out, tok)
		}

		i++
	}

	return out, diags
}

// syntheticToken creates a synthesized INDENT or DEDENT token at the position
// of the given reference token. Len=0 marks it as synthesized.
func syntheticToken(kind TokenKind, refTok Token) Token {
	return Token{Kind: kind, Offset: refTok.Offset, Len: 0}
}

// countLeadingSpaces counts consecutive space bytes (0x20) starting at src[offset].
func countLeadingSpaces(src []byte, offset int) int {
	count := 0
	for i := offset; i < len(src); i++ {
		if src[i] == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// findNextContentOffset scans forward through the token stream starting at startIdx
// to find the byte offset of the first token on a non-blank, non-comment-only line.
// Returns -1 if no content remains (only NEWLINEs and EOF left).
func findNextContentOffset(tokens []Token, startIdx int, src []byte) int {
	for j := startIdx; j < len(tokens); j++ {
		tok := tokens[j]
		if tok.Kind == TokenNewline {
			continue // skip blank lines
		}
		if tok.Kind == TokenEOF {
			return -1 // no more content
		}
		// Found a real token — find the start of its line
		lineStart := findLineStart(src, int(tok.Offset))
		return lineStart
	}
	return -1
}

// findLineStart returns the byte offset of the start of the line containing pos.
func findLineStart(src []byte, pos int) int {
	for i := pos - 1; i >= 0; i-- {
		if src[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

// indentStack tracks indentation levels.
type indentStack struct {
	levels []int
}

func (s *indentStack) top() int {
	return s.levels[len(s.levels)-1]
}

func (s *indentStack) push(n int) {
	s.levels = append(s.levels, n)
}

func (s *indentStack) pop() {
	s.levels = s.levels[:len(s.levels)-1]
}

// formatIndentError formats an indentation error message.
func formatIndentError(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
