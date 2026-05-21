package lexer

import (
	"fmt"
	"sort"

	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

// Lex tokenizes the UTF-8 source into a flat token slice.
// src must remain alive for the lifetime of the returned tokens
// (tokens reference into src by offset/length, no copying).
// Returns the token slice, a line table for error reporting, and any diagnostics.
// Lexing always completes — errors produce TokenError tokens and diagnostics,
// but the returned slice is always non-nil.
func Lex(src []byte) ([]Token, *LineTable, []diagnostics.Diagnostic) {
	l := &lexer{
		src:    src,
		tokens: make([]Token, 0, len(src)/4+16),
		lt:     LineTable{srcLen: uint32(len(src))},
	}
	l.run()
	processed, indentDiags := ProcessIndentation(src, l.tokens, &l.lt)
	allDiags := append(l.diags, indentDiags...)
	return processed, &l.lt, allDiags
}

// LineTable records the byte offset of each newline in the source.
// newlineOffsets[i] is the offset of the '\n' at the end of line i+1.
// Line numbers are 1-based; column numbers are 1-based.
type LineTable struct {
	newlineOffsets []uint32 // sorted, one entry per \n in source
	srcLen         uint32
}

// LineCol returns the 1-based line and column for a byte offset.
func (lt *LineTable) LineCol(offset uint32) (line, col uint32) {
	lo := sort.Search(len(lt.newlineOffsets), func(i int) bool {
		return lt.newlineOffsets[i] >= offset
	})
	line = uint32(lo) + 1
	lineStart := uint32(0)
	if lo > 0 {
		lineStart = lt.newlineOffsets[lo-1] + 1
	}
	col = offset - lineStart + 1
	return
}

// lexer is the internal scanning state. Unexported — callers use Lex().
type lexer struct {
	src    []byte
	pos    int // current byte position
	tokens []Token
	lt     LineTable
	diags  []diagnostics.Diagnostic
}

// run is the main dispatch loop.
func (l *lexer) run() {
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		switch {
		case b == ' ' || b == '\r':
			l.pos++
		case b == '@':
			if l.pos+1 < len(l.src) && isIdentStart(l.src[l.pos+1]) {
				l.pos++
			} else {
				l.scanOperatorOrPunct()
			}
		case b == '\t':
			l.addDiag(1, "tab character not allowed; use 4 spaces for indentation")
			l.pos++
		case b == '\n':
			l.emitNewline()
		case b == '/' && l.peek1() == '/':
			l.scanLineComment()
		case b >= '0' && b <= '9':
			l.scanNumber()
		case b == '"':
			l.scanString()
		case b == '\'':
			l.scanChar()
		case isIdentStart(b):
			l.scanIdent()
		default:
			l.scanOperatorOrPunct()
		}
	}
	l.emit(TokenEOF, uint32(l.pos), 0)
}

// emit appends a token to the output slice.
func (l *lexer) emit(kind TokenKind, offset uint32, length uint16) {
	l.tokens = append(l.tokens, Token{Kind: kind, Offset: offset, Len: length})
}

// emitNewline records the newline in the line table and emits a NEWLINE token.
func (l *lexer) emitNewline() {
	offset := uint32(l.pos)
	l.lt.newlineOffsets = append(l.lt.newlineOffsets, offset)
	l.emit(TokenNewline, offset, 1)
	l.pos++
}

// addDiag adds a diagnostic at the current position.
func (l *lexer) addDiag(code uint32, msg string) {
	line, col := uint32(1), uint32(l.pos+1)
	if len(l.lt.newlineOffsets) > 0 {
		line, col = l.lt.LineCol(uint32(l.pos))
	}
	l.diags = append(l.diags, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     code,
		Pos:      diagnostics.Pos{Offset: uint32(l.pos), Line: line, Col: col},
		Message:  msg,
	})
}

// emitErrorToken emits a TokenError and records a diagnostic.
func (l *lexer) emitErrorToken(code uint32, msg string) {
	l.addDiag(code, msg)
	l.emit(TokenError, uint32(l.pos), 1)
}

// peek1 looks at the next byte without advancing.
func (l *lexer) peek1() byte {
	if l.pos+1 < len(l.src) {
		return l.src[l.pos+1]
	}
	return 0
}

// scanLineComment advances past a // comment until end of line.
func (l *lexer) scanLineComment() {
	l.pos += 2 // skip //
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
	// Don't consume the \n — the main loop handles it
}

// scanNumber scans all integer and float literal forms.
func (l *lexer) scanNumber() {
	start := l.pos
	b := l.src[l.pos]

	// Check for hex/octal/binary prefix
	if b == '0' && l.pos+1 < len(l.src) {
		next := l.src[l.pos+1]
		switch next {
		case 'x', 'X':
			l.pos += 2
			l.scanHexDigits()
			l.emit(TokenIntLit, uint32(start), uint16(l.pos-start))
			return
		case 'o', 'O':
			l.pos += 2
			l.scanOctDigits()
			l.emit(TokenIntLit, uint32(start), uint16(l.pos-start))
			return
		case 'b', 'B':
			l.pos += 2
			l.scanBinDigits()
			l.emit(TokenIntLit, uint32(start), uint16(l.pos-start))
			return
		}
	}

	// Decimal integer or float
	l.scanDecDigits()

	// Check for float: decimal point followed by digit
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		// Peek ahead: if next char is a digit, it's a float
		if l.pos+1 < len(l.src) && l.src[l.pos+1] >= '0' && l.src[l.pos+1] <= '9' {
			l.pos++ // consume '.'
			l.scanDecDigits()
			// Optional exponent
			if l.pos < len(l.src) && (l.src[l.pos] == 'e' || l.src[l.pos] == 'E') {
				l.pos++
				if l.pos < len(l.src) && (l.src[l.pos] == '+' || l.src[l.pos] == '-') {
					l.pos++
				}
				l.scanDecDigits()
			}
			l.emit(TokenFloatLit, uint32(start), uint16(l.pos-start))
			return
		}
	}

	l.emit(TokenIntLit, uint32(start), uint16(l.pos-start))
}

func (l *lexer) scanDecDigits() {
	for l.pos < len(l.src) && (isDecDigit(l.src[l.pos]) || l.src[l.pos] == '_') {
		l.pos++
	}
}

func (l *lexer) scanHexDigits() {
	for l.pos < len(l.src) && (isHexDigit(l.src[l.pos]) || l.src[l.pos] == '_') {
		l.pos++
	}
}

func (l *lexer) scanOctDigits() {
	for l.pos < len(l.src) && (isOctDigit(l.src[l.pos]) || l.src[l.pos] == '_') {
		l.pos++
	}
}

func (l *lexer) scanBinDigits() {
	for l.pos < len(l.src) && (l.src[l.pos] == '0' || l.src[l.pos] == '1' || l.src[l.pos] == '_') {
		l.pos++
	}
}

// scanString scans a double-quoted string literal.
func (l *lexer) scanString() {
	start := l.pos
	l.pos++ // skip opening '"'

	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b == '"' {
			l.pos++ // skip closing '"'
			l.emit(TokenStringLit, uint32(start), uint16(l.pos-start))
			return
		}
		if b == '\\' {
			l.pos++ // skip backslash
			if l.pos < len(l.src) {
				l.pos++ // skip escaped character
			}
			continue
		}
		if b == '\n' {
			// Unterminated string at newline
			l.addDiag(2, "unterminated string literal")
			l.emit(TokenStringLit, uint32(start), uint16(l.pos-start))
			return
		}
		l.pos++
	}

	// EOF in string
	l.addDiag(2, "unterminated string literal at end of file")
	l.emit(TokenStringLit, uint32(start), uint16(l.pos-start))
}

// scanChar scans a single-quoted character literal.
func (l *lexer) scanChar() {
	start := l.pos
	l.pos++ // skip opening '\''

	if l.pos < len(l.src) {
		if l.src[l.pos] == '\\' {
			l.pos++ // skip backslash
			if l.pos < len(l.src) {
				l.pos++ // skip escaped char
			}
		} else if l.src[l.pos] != '\'' {
			l.pos++ // skip the character
		}
	}

	if l.pos < len(l.src) && l.src[l.pos] == '\'' {
		l.pos++ // skip closing '\''
	} else {
		l.addDiag(3, "unterminated character literal")
	}

	l.emit(TokenCharLit, uint32(start), uint16(l.pos-start))
}

// scanIdent scans an identifier or keyword.
func (l *lexer) scanIdent() {
	start := l.pos
	for l.pos < len(l.src) && isIdentContinue(l.src[l.pos]) {
		l.pos++
	}
	text := string(l.src[start:l.pos])
	kind := TokenIdent
	if kw, ok := Keywords[text]; ok {
		kind = kw
	}
	l.emit(kind, uint32(start), uint16(l.pos-start))
}

// scanOperatorOrPunct scans multi-character operators before single-character ones.
func (l *lexer) scanOperatorOrPunct() {
	start := l.pos
	b := l.src[l.pos]

	// Try two-character operators first
	if l.pos+1 < len(l.src) {
		two := string(l.src[l.pos : l.pos+2])
		if kind, ok := twoCharOps[two]; ok {
			l.pos += 2
			l.emit(kind, uint32(start), 2)
			return
		}
	}

	// Single-character operator or punctuation
	if kind, ok := oneCharOps[b]; ok {
		l.pos++
		l.emit(kind, uint32(start), 1)
		return
	}

	// Unknown character
	l.emitErrorToken(4, fmt.Sprintf("unexpected character %q (0x%02x)", rune(b), b))
	l.pos++
}

// isIdentStart reports whether b can start an identifier (ASCII only for MVP).
func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// isIdentContinue reports whether b can continue an identifier.
func isIdentContinue(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

func isDecDigit(b byte) bool { return b >= '0' && b <= '9' }

func isHexDigit(b byte) bool {
	return isDecDigit(b) || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func isOctDigit(b byte) bool { return b >= '0' && b <= '7' }

// Two-character operator lookup table.
var twoCharOps = map[string]TokenKind{
	"==": TokenEqEq, "!=": TokenBangEq, "<=": TokenLtEq, ">=": TokenGtEq,
	"**": TokenStarStar, "<<": TokenLtLt, ">>": TokenGtGt,
	"->": TokenArrow, ":=": TokenColonEq,
	"+=": TokenPlusEq, "-=": TokenMinusEq, "*=": TokenStarEq,
	"/=": TokenSlashEq, "%=": TokenPercentEq, ".*": TokenDotStar, "..": TokenDotDot,
}

// Single-character operator/punctuation lookup table.
var oneCharOps = map[byte]TokenKind{
	'+': TokenPlus, '-': TokenMinus, '*': TokenStar, '/': TokenSlash,
	'%': TokenPercent, '=': TokenEq, '<': TokenLt, '>': TokenGt,
	'&': TokenAmp, '|': TokenPipe, '^': TokenCaret, '~': TokenTilde,
	'.': TokenDot, ',': TokenComma, ':': TokenColon, ';': TokenSemicolon,
	'!': TokenBang, '(': TokenLParen, ')': TokenRParen,
	'[': TokenLBracket, ']': TokenRBracket, '{': TokenLBrace, '}': TokenRBrace,
}
