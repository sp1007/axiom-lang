package lexer

import "testing"

// tokenKinds extracts all token kinds from a slice
func tokenKinds(toks []Token) []TokenKind {
	kinds := make([]TokenKind, len(toks))
	for i, tok := range toks {
		kinds[i] = tok.Kind
	}
	return kinds
}

// countKind counts tokens of a specific kind
func countKind(toks []Token, kind TokenKind) int {
	n := 0
	for _, tok := range toks {
		if tok.Kind == kind {
			n++
		}
	}
	return n
}

// containsKind checks if any token has the given kind
func containsKind(toks []Token, kind TokenKind) bool {
	return countKind(toks, kind) > 0
}

func TestIndentDedentBasic(t *testing.T) {
	// fn main():\n    let x = 1\n
	src := "fn main():\n    let x = 1\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	assertf(t, containsKind(toks, TokenIndent), "expected INDENT in output")
	assertf(t, containsKind(toks, TokenDedent), "expected DEDENT in output")
}

func TestIndentNestedBlocks(t *testing.T) {
	src := "if x:\n    if y:\n        let z = 1\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	indents := countKind(toks, TokenIndent)
	dedents := countKind(toks, TokenDedent)
	assertf(t, indents == 2, "expected 2 INDENTs, got %d", indents)
	assertf(t, dedents == 2, "expected 2 DEDENTs, got %d", dedents)
}

func TestIndentBlankLinesIgnored(t *testing.T) {
	src := "fn main():\n\n    let x = 1\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	assertf(t, containsKind(toks, TokenIndent), "expected INDENT despite blank line")
}

func TestIndentBadAmount(t *testing.T) {
	// 2-space indent is invalid
	src := "fn main():\n  let x = 1\n"
	_, _, diags := Lex([]byte(src))
	found := false
	for _, d := range diags {
		if d.Code == 10 {
			found = true
		}
	}
	assertf(t, found, "expected E0010 diagnostic for non-4-space indent")
}

func TestDedentMismatch(t *testing.T) {
	// Dedenting to 3 spaces which doesn't match any enclosing level
	src := "if x:\n    if y:\n        let z = 1\n   let w = 2\n"
	_, _, diags := Lex([]byte(src))
	found := false
	for _, d := range diags {
		if d.Code == 11 {
			found = true
		}
	}
	assertf(t, found, "expected E0011 diagnostic for mismatched dedent")
}

func TestIndentEOFDedentsEmitted(t *testing.T) {
	src := "fn main():\n    if x:\n        y()\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	dedents := countKind(toks, TokenDedent)
	assertf(t, dedents == 2, "expected 2 EOF DEDENTs, got %d", dedents)
}

func TestIndentSynthesizedTokensHaveZeroLen(t *testing.T) {
	src := "fn main():\n    x()\n"
	toks, _, _ := Lex([]byte(src))
	for _, tok := range toks {
		if tok.Kind == TokenIndent || tok.Kind == TokenDedent {
			assertf(t, tok.Len == 0, "synthesized %s token has Len=%d, want 0", tok.Kind, tok.Len)
		}
	}
}

func TestIndentCommentOnlyLineIgnored(t *testing.T) {
	src := "fn main():\n    // comment\n    let x = 1\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	indents := countKind(toks, TokenIndent)
	assertf(t, indents == 1, "expected 1 INDENT, got %d", indents)
}

func TestIndentBalance(t *testing.T) {
	// Every INDENT must be matched by a DEDENT
	src := "fn foo():\n    if x:\n        a()\n    b()\nc()\n"
	toks, _, _ := Lex([]byte(src))
	indents := countKind(toks, TokenIndent)
	dedents := countKind(toks, TokenDedent)
	assertf(t, indents == dedents, "INDENT/DEDENT mismatch: %d INDENTs vs %d DEDENTs", indents, dedents)
}

func TestIndentDedentMultipleLevelsAtOnce(t *testing.T) {
	// Drop from level 8 to level 0 in one step
	src := "if a:\n    if b:\n        x()\ny()\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	dedents := countKind(toks, TokenDedent)
	assertf(t, dedents == 2, "expected 2 DEDENTs at once, got %d", dedents)
}

func TestIndentFlatCode(t *testing.T) {
	// No indentation — should have no INDENT/DEDENT
	src := "let x = 1\nlet y = 2\nlet z = 3\n"
	toks, _, diags := Lex([]byte(src))
	requireNoErrors(t, diags)
	indents := countKind(toks, TokenIndent)
	dedents := countKind(toks, TokenDedent)
	assertf(t, indents == 0, "expected 0 INDENTs for flat code, got %d", indents)
	assertf(t, dedents == 0, "expected 0 DEDENTs for flat code, got %d", dedents)
}

func TestIndentTokenOrder(t *testing.T) {
	// INDENT must come after NEWLINE
	src := "fn main():\n    x()\n"
	toks, _, _ := Lex([]byte(src))
	for i, tok := range toks {
		if tok.Kind == TokenIndent {
			assertf(t, i > 0, "INDENT at index 0, expected after NEWLINE")
			assertf(t, toks[i-1].Kind == TokenNewline, "expected NEWLINE before INDENT, got %s", toks[i-1].Kind)
			break
		}
	}
}

func TestIndentEOFAlwaysLast(t *testing.T) {
	src := "fn main():\n    x()\n"
	toks, _, _ := Lex([]byte(src))
	assertf(t, len(toks) > 0, "expected non-empty token stream")
	assertf(t, toks[len(toks)-1].Kind == TokenEOF, "expected EOF as last token, got %s", toks[len(toks)-1].Kind)
}
