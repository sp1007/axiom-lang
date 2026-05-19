package lexer

import "testing"

func TestTokenKindFitsUint8(t *testing.T) {
	if TokenKindCount > 255 {
		t.Fatalf("TokenKindCount = %d > 255; TokenKind must fit in uint8", TokenKindCount)
	}
}

func TestTokenKindStringNonEmpty(t *testing.T) {
	// Every defined kind should have a meaningful String() representation
	for i := TokenKind(0); i < TokenKindCount; i++ {
		s := i.String()
		if s == "" {
			t.Errorf("TokenKind(%d).String() is empty; add to tokenKindNames", i)
		}
	}
}

func TestKeywordsMapComplete(t *testing.T) {
	// Every keyword token must appear in the Keywords map
	kwTokens := []TokenKind{
		TokenAnd, TokenAs, TokenAsync, TokenAwait, TokenConst, TokenDefer,
		TokenElif, TokenElse, TokenExtern, TokenFalse, TokenFn,
		TokenFor, TokenIf, TokenImport, TokenIn, TokenInterface,
		TokenIsolated, TokenFuture, TokenLent, TokenLet, TokenMatch,
		TokenMut, TokenNil, TokenNot, TokenOr, TokenPacked,
		TokenPub, TokenReturn, TokenSpawn, TokenStruct, TokenTrue,
		TokenType, TokenUnsafe, TokenWhile,
	}
	inMap := map[TokenKind]bool{}
	for _, v := range Keywords {
		inMap[v] = true
	}
	for _, kw := range kwTokens {
		if !inMap[kw] {
			t.Errorf("keyword token %s (kind=%d) not found in Keywords map", kw, kw)
		}
	}
}

func TestKeywordsMapNoDuplicateValues(t *testing.T) {
	seen := map[TokenKind]string{}
	for text, kind := range Keywords {
		if prev, ok := seen[kind]; ok {
			t.Errorf("TokenKind %d mapped to both %q and %q", kind, prev, text)
		}
		seen[kind] = text
	}
}

func TestKeywordIsNotIdent(t *testing.T) {
	// All keywords must NOT be TokenIdent
	for text, kind := range Keywords {
		if kind == TokenIdent {
			t.Errorf("keyword %q maps to TokenIdent; must have its own kind", text)
		}
	}
}

func TestTokenKindIsKeyword(t *testing.T) {
	if !TokenFn.IsKeyword() {
		t.Error("TokenFn should be a keyword")
	}
	if !TokenWhile.IsKeyword() {
		t.Error("TokenWhile should be a keyword")
	}
	if TokenIdent.IsKeyword() {
		t.Error("TokenIdent should not be a keyword")
	}
	if TokenPlus.IsKeyword() {
		t.Error("TokenPlus should not be a keyword")
	}
}

func TestTokenKindIsLiteral(t *testing.T) {
	if !TokenIntLit.IsLiteral() {
		t.Error("TokenIntLit should be a literal")
	}
	if !TokenStringLit.IsLiteral() {
		t.Error("TokenStringLit should be a literal")
	}
	if TokenIdent.IsLiteral() {
		t.Error("TokenIdent should not be a literal")
	}
}

func TestTokenKindIsOperator(t *testing.T) {
	if !TokenPlus.IsOperator() {
		t.Error("TokenPlus should be an operator")
	}
	if !TokenEqEq.IsOperator() {
		t.Error("TokenEqEq should be an operator")
	}
	if TokenIdent.IsOperator() {
		t.Error("TokenIdent should not be an operator")
	}
}

func TestTokenKindStringKnown(t *testing.T) {
	// Spot-check some well-known kinds
	tests := []struct {
		kind TokenKind
		want string
	}{
		{TokenIntLit, "integer literal"},
		{TokenFn, "'fn'"},
		{TokenPlus, "'+'"},
		{TokenEOF, "EOF"},
		{TokenError, "ERROR"},
		{TokenIndent, "INDENT"},
		{TokenArrow, "'->'"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestTokenKindStringUnknown(t *testing.T) {
	// Out-of-range kind should return "TokenKind(N)"
	unknown := TokenKind(255)
	s := unknown.String()
	if s == "" {
		t.Error("unknown TokenKind.String() should not be empty")
	}
}
