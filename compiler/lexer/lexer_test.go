package lexer

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

// Helper: fail if there are any error-level diagnostics
func requireNoErrors(t *testing.T, diags []diagnostics.Diagnostic) {
	t.Helper()
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			t.Fatalf("unexpected error: %s", d.Message)
		}
	}
}

// Helper: assert condition with optional format message
func assertf(t *testing.T, cond bool, format string, args ...any) {
	t.Helper()
	if !cond {
		t.Fatalf(format, args...)
	}
}

// Helper: extract token text from source
func tokenText(src []byte, tok Token) string {
	return string(src[tok.Offset : tok.Offset+uint32(tok.Len)])
}

// ========== Empty / EOF tests ==========

func TestLexerEmpty(t *testing.T) {
	toks, _, diags := Lex([]byte{})
	requireNoErrors(t, diags)
	assertf(t, len(toks) == 1, "expected 1 token (EOF), got %d", len(toks))
	assertf(t, toks[0].Kind == TokenEOF, "expected EOF, got %s", toks[0].Kind)
}

func TestLexerWhitespaceOnly(t *testing.T) {
	toks, _, diags := Lex([]byte("   "))
	requireNoErrors(t, diags)
	assertf(t, toks[len(toks)-1].Kind == TokenEOF, "expected EOF at end")
}

// ========== Integer literal tests ==========

func TestLexerIntLitDecimal(t *testing.T) {
	src := []byte("42")
	toks, _, diags := Lex(src)
	requireNoErrors(t, diags)
	assertf(t, toks[0].Kind == TokenIntLit, "expected IntLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Offset == 0, "expected offset 0, got %d", toks[0].Offset)
	assertf(t, toks[0].Len == 2, "expected len 2, got %d", toks[0].Len)
}

func TestLexerIntLitHex(t *testing.T) {
	src := []byte("0xFF")
	toks, _, diags := Lex(src)
	requireNoErrors(t, diags)
	assertf(t, toks[0].Kind == TokenIntLit, "expected IntLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 4, "expected len 4, got %d", toks[0].Len)
}

func TestLexerIntLitBinary(t *testing.T) {
	src := []byte("0b1010_1010")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenIntLit, "expected IntLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 11, "expected len 11, got %d", toks[0].Len)
}

func TestLexerIntLitOctal(t *testing.T) {
	src := []byte("0o755")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenIntLit, "expected IntLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 5, "expected len 5, got %d", toks[0].Len)
}

func TestLexerIntLitWithUnderscores(t *testing.T) {
	src := []byte("1_000_000")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenIntLit, "expected IntLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 9, "expected len 9, got %d", toks[0].Len)
}

// ========== Float literal tests ==========

func TestLexerFloatLit(t *testing.T) {
	src := []byte("3.14")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenFloatLit, "expected FloatLit, got %s", toks[0].Kind)
}

func TestLexerFloatLitExponent(t *testing.T) {
	src := []byte("1.0e-6")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenFloatLit, "expected FloatLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 6, "expected len 6, got %d", toks[0].Len)
}

func TestLexerFloatLitPositiveExponent(t *testing.T) {
	src := []byte("2.5E+10")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenFloatLit, "expected FloatLit, got %s", toks[0].Kind)
}

func TestLexerIntFollowedByDot(t *testing.T) {
	// 42.foo should be: IntLit(42), Dot, Ident(foo)
	src := []byte("42.foo")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenIntLit, "expected IntLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 2, "expected len 2, got %d", toks[0].Len)
	assertf(t, toks[1].Kind == TokenDot, "expected Dot, got %s", toks[1].Kind)
	assertf(t, toks[2].Kind == TokenIdent, "expected Ident, got %s", toks[2].Kind)
}

// ========== String literal tests ==========

func TestLexerStringLit(t *testing.T) {
	src := []byte(`"hello"`)
	toks, _, diags := Lex(src)
	requireNoErrors(t, diags)
	assertf(t, toks[0].Kind == TokenStringLit, "expected StringLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 7, "expected len 7, got %d", toks[0].Len)
}

func TestLexerStringEscapes(t *testing.T) {
	src := []byte(`"\n\t\\\""`)
	toks, _, diags := Lex(src)
	requireNoErrors(t, diags)
	assertf(t, toks[0].Kind == TokenStringLit, "expected StringLit, got %s", toks[0].Kind)
}

func TestLexerStringUnterminated(t *testing.T) {
	src := []byte(`"hello`)
	toks, _, diags := Lex(src)
	assertf(t, len(diags) > 0, "expected diagnostic for unterminated string")
	assertf(t, toks[0].Kind == TokenStringLit, "expected StringLit even on error")
}

func TestLexerStringNewline(t *testing.T) {
	src := []byte("\"hello\nworld\"")
	_, _, diags := Lex(src)
	assertf(t, len(diags) > 0, "expected diagnostic for newline in string")
}

// ========== Char literal tests ==========

func TestLexerCharLit(t *testing.T) {
	src := []byte("'a'")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenCharLit, "expected CharLit, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 3, "expected len 3, got %d", toks[0].Len)
}

func TestLexerCharEscape(t *testing.T) {
	src := []byte("'\\n'")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenCharLit, "expected CharLit, got %s", toks[0].Kind)
}

// ========== Keyword tests ==========

func TestLexerAllKeywords(t *testing.T) {
	for text, expectedKind := range Keywords {
		toks, _, _ := Lex([]byte(text))
		if toks[0].Kind != expectedKind {
			t.Errorf("keyword %q: got %s, want %s", text, toks[0].Kind, expectedKind)
		}
	}
}

func TestLexerIdentNotKeyword(t *testing.T) {
	toks, _, _ := Lex([]byte("foobar"))
	assertf(t, toks[0].Kind == TokenIdent, "expected Ident, got %s", toks[0].Kind)
}

func TestLexerIdentWithDigits(t *testing.T) {
	toks, _, _ := Lex([]byte("foo123"))
	assertf(t, toks[0].Kind == TokenIdent, "expected Ident, got %s", toks[0].Kind)
	assertf(t, toks[0].Len == 6, "expected len 6, got %d", toks[0].Len)
}

func TestLexerIdentUnderscore(t *testing.T) {
	toks, _, _ := Lex([]byte("_private"))
	assertf(t, toks[0].Kind == TokenIdent, "expected Ident, got %s", toks[0].Kind)
}

// ========== Operator tests ==========

func TestLexerAllOperators(t *testing.T) {
	cases := []struct {
		src  string
		want TokenKind
	}{
		{"==", TokenEqEq}, {"!=", TokenBangEq}, {"<=", TokenLtEq},
		{">=", TokenGtEq}, {"**", TokenStarStar}, {"<<", TokenLtLt},
		{">>", TokenGtGt}, {"->", TokenArrow}, {":=", TokenColonEq},
		{"+=", TokenPlusEq}, {"-=", TokenMinusEq}, {"*=", TokenStarEq},
		{"/=", TokenSlashEq}, {"%=", TokenPercentEq}, {".*", TokenDotStar},
		{"+", TokenPlus}, {"-", TokenMinus}, {"*", TokenStar}, {"/", TokenSlash},
		{"%", TokenPercent}, {"=", TokenEq}, {"<", TokenLt}, {">", TokenGt},
		{"&", TokenAmp}, {"|", TokenPipe}, {"^", TokenCaret}, {"~", TokenTilde},
		{".", TokenDot}, {",", TokenComma}, {":", TokenColon}, {";", TokenSemicolon},
		{"!", TokenBang}, {"(", TokenLParen}, {")", TokenRParen},
		{"[", TokenLBracket}, {"]", TokenRBracket}, {"{", TokenLBrace}, {"}", TokenRBrace},
	}
	for _, c := range cases {
		toks, _, _ := Lex([]byte(c.src))
		if toks[0].Kind != c.want {
			t.Errorf("%q: got %s, want %s", c.src, toks[0].Kind, c.want)
		}
	}
}

// ========== Comment tests ==========

func TestLexerLineComment(t *testing.T) {
	toks, _, _ := Lex([]byte("// this is a comment\nfoo"))
	assertf(t, toks[0].Kind == TokenNewline, "expected NEWLINE, got %s", toks[0].Kind)
	assertf(t, toks[1].Kind == TokenIdent, "expected Ident, got %s", toks[1].Kind)
}

func TestLexerCommentAtEOF(t *testing.T) {
	toks, _, _ := Lex([]byte("// comment only"))
	assertf(t, toks[0].Kind == TokenEOF, "expected EOF, got %s", toks[0].Kind)
}

// ========== LineTable tests ==========

func TestLexerLineTable(t *testing.T) {
	src := []byte("foo\nbar\nbaz")
	_, lt, _ := Lex(src)
	line, col := lt.LineCol(4) // 'b' of 'bar'
	assertf(t, line == 2, "expected line 2, got %d", line)
	assertf(t, col == 1, "expected col 1, got %d", col)
}

func TestLexerLineTableFirstLine(t *testing.T) {
	src := []byte("foo\nbar")
	_, lt, _ := Lex(src)
	line, col := lt.LineCol(0) // 'f' of 'foo'
	assertf(t, line == 1, "expected line 1, got %d", line)
	assertf(t, col == 1, "expected col 1, got %d", col)
}

func TestLexerLineTableLastLine(t *testing.T) {
	src := []byte("a\nb\nc")
	_, lt, _ := Lex(src)
	line, col := lt.LineCol(4) // 'c'
	assertf(t, line == 3, "expected line 3, got %d", line)
	assertf(t, col == 1, "expected col 1, got %d", col)
}

// ========== Newline tests ==========

func TestLexerNewlineTokens(t *testing.T) {
	src := []byte("a\nb")
	toks, _, _ := Lex(src)
	assertf(t, toks[0].Kind == TokenIdent, "expected Ident, got %s", toks[0].Kind)
	assertf(t, toks[1].Kind == TokenNewline, "expected NEWLINE, got %s", toks[1].Kind)
	assertf(t, toks[2].Kind == TokenIdent, "expected Ident, got %s", toks[2].Kind)
	assertf(t, toks[3].Kind == TokenEOF, "expected EOF, got %s", toks[3].Kind)
}

// ========== Tab error tests ==========

func TestLexerTabError(t *testing.T) {
	_, _, diags := Lex([]byte("\t"))
	assertf(t, len(diags) > 0, "expected diagnostic for tab")
	assertf(t, diags[0].Code == 1, "expected E0001, got E%04d", diags[0].Code)
}

// ========== Unknown character tests ==========

func TestLexerUnknownChar(t *testing.T) {
	toks, _, diags := Lex([]byte("@"))
	assertf(t, len(diags) > 0, "expected diagnostic for unknown char")
	assertf(t, toks[0].Kind == TokenError, "expected Error token, got %s", toks[0].Kind)
}

// ========== Multi-token sequence tests ==========

func TestLexerFuncDecl(t *testing.T) {
	src := []byte("fn main():")
	toks, _, diags := Lex(src)
	requireNoErrors(t, diags)

	expected := []TokenKind{
		TokenFn, TokenIdent, TokenLParen, TokenRParen, TokenColon, TokenEOF,
	}
	for i, want := range expected {
		if i >= len(toks) {
			t.Fatalf("expected %d tokens, got %d", len(expected), len(toks))
		}
		if toks[i].Kind != want {
			t.Errorf("token[%d]: got %s, want %s", i, toks[i].Kind, want)
		}
	}
}

func TestLexerLetDecl(t *testing.T) {
	src := []byte("let x: i32 = 42")
	toks, _, diags := Lex(src)
	requireNoErrors(t, diags)

	expected := []TokenKind{
		TokenLet, TokenIdent, TokenColon, TokenIdent, TokenEq, TokenIntLit, TokenEOF,
	}
	for i, want := range expected {
		if i >= len(toks) {
			t.Fatalf("expected %d tokens, got %d", len(expected), len(toks))
		}
		if toks[i].Kind != want {
			t.Errorf("token[%d]: got %s, want %s", i, toks[i].Kind, want)
		}
	}
}

func TestLexerReturnArrow(t *testing.T) {
	src := []byte("fn foo() -> i32:")
	toks, _, _ := Lex(src)

	expected := []TokenKind{
		TokenFn, TokenIdent, TokenLParen, TokenRParen, TokenArrow, TokenIdent, TokenColon, TokenEOF,
	}
	for i, want := range expected {
		if i >= len(toks) {
			t.Fatalf("expected %d tokens, got %d", len(expected), len(toks))
		}
		if toks[i].Kind != want {
			t.Errorf("token[%d]: got %s, want %s", i, toks[i].Kind, want)
		}
	}
}

func TestLexerZeroCopyProperty(t *testing.T) {
	src := []byte("hello world 42")
	toks, _, _ := Lex(src)
	// Verify tokens reference into the original source
	for _, tok := range toks {
		if tok.Kind == TokenEOF {
			continue
		}
		text := tokenText(src, tok)
		if len(text) == 0 && tok.Len > 0 {
			t.Errorf("token %s at offset %d: empty text but len=%d", tok.Kind, tok.Offset, tok.Len)
		}
	}
}

func TestLexerNeverPanics(t *testing.T) {
	// Feed various edge cases, none should panic
	inputs := []string{
		"", "\n", "\t", "'", "\"", "'\\", "\"\\", "0x", "0b", "0o",
		"//", "/*", "@#$", "\x00", "\xff",
	}
	for _, input := range inputs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Lex(%q) panicked: %v", input, r)
				}
			}()
			Lex([]byte(input))
		}()
	}
}
