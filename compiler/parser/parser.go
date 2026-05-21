// Package parser implements the AXIOM recursive descent parser.
// It transforms a post-processed token stream (with INDENT/DEDENT/NEWLINE tokens)
// into a flat AstTree. The parser never panics; all errors produce diagnostics.
package parser

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

// Parser converts a token stream into an AstTree.
// Single-pass recursive descent following the EBNF grammar.
// Never panics; errors produce diagnostics and NodeError nodes.
type Parser struct {
	tokens []lexer.Token
	pos    int
	tree   *ast.AstTree
	pool   *ast.InternPool
	src    []byte
	diags  []diagnostics.Diagnostic
}

// Parse parses the token stream and returns an AstTree.
// src is the original source bytes (for error reporting and token text recovery).
// Always returns a non-nil tree; errors are in the returned diagnostics.
func Parse(tokens []lexer.Token, src []byte, pool *ast.InternPool) (*ast.AstTree, []diagnostics.Diagnostic) {
	p := &Parser{
		tokens: tokens,
		tree:   ast.NewTree(src, tokens),
		pool:   pool,
		src:    src,
	}
	p.parseProgram()
	return p.tree, p.diags
}

// =============================================================================
// Core helpers
// =============================================================================

// peek returns the current token without advancing, skipping bare NEWLINEs.
// Side-effect: advances p.pos past any leading NEWLINE tokens.
func (p *Parser) peek() lexer.Token {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Kind == lexer.TokenNewline {
		p.pos++
	}
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

// peekRaw returns the current token without skipping NEWLINEs.
func (p *Parser) peekRaw() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

// consume advances past the current token (skipping leading NEWLINEs) and returns it.
func (p *Parser) consume() lexer.Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

// check returns true if the current non-NEWLINE token has the given kind.
func (p *Parser) check(kind lexer.TokenKind) bool {
	return p.peek().Kind == kind
}

// checkRaw returns true if the raw current token (no NEWLINE skip) has the given kind.
func (p *Parser) checkRaw(kind lexer.TokenKind) bool {
	return p.peekRaw().Kind == kind
}

// expect consumes a token of the given kind or emits a diagnostic.
// On mismatch, the token is NOT consumed (caller may still proceed).
func (p *Parser) expect(kind lexer.TokenKind) (lexer.Token, bool) {
	tok := p.peek()
	if tok.Kind != kind {
		p.errorf(tok, "expected %s, got %s", kind, tok.Kind)
		return tok, false
	}
	return p.consume(), true
}

// expectNewline consumes the immediately-following NEWLINE token.
// Does not emit an error if we are at EOF or DEDENT (natural block boundary).
func (p *Parser) expectNewline() {
	if p.pos < len(p.tokens) && p.tokens[p.pos].Kind == lexer.TokenNewline {
		p.pos++
		return
	}
	raw := p.peekRaw()
	if raw.Kind != lexer.TokenEOF && raw.Kind != lexer.TokenDedent {
		p.errorf(raw, "expected newline, got %s", raw.Kind)
	}
}

// tokenIdx returns the index of tok in the token slice using binary search on Offset.
func (p *Parser) tokenIdx(tok lexer.Token) uint32 {
	lo, hi := 0, len(p.tokens)
	for lo < hi {
		mid := (lo + hi) / 2
		if p.tokens[mid].Offset < tok.Offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(p.tokens) {
		return uint32(lo)
	}
	if len(p.tokens) > 0 {
		return uint32(len(p.tokens) - 1)
	}
	return 0
}

// tokenText returns the source bytes for tok. Returns nil if src is nil or tok is synthetic.
func (p *Parser) tokenText(tok lexer.Token) []byte {
	if p.src == nil || tok.Len == 0 {
		return nil
	}
	end := tok.Offset + uint32(tok.Len)
	if end > uint32(len(p.src)) {
		return nil
	}
	return p.src[tok.Offset:end]
}

// =============================================================================
// Error handling and recovery
// =============================================================================

func (p *Parser) errorf(tok lexer.Token, format string, args ...any) {
	p.diags = append(p.diags, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     1000,
		Pos:      diagnostics.Pos{Offset: tok.Offset},
		Message:  fmt.Sprintf(format, args...),
	})
}

// syncToTopLevel advances until a top-level declaration keyword is found.
func (p *Parser) syncToTopLevel() {
	for !p.check(lexer.TokenEOF) {
		switch p.peek().Kind {
		case lexer.TokenFn, lexer.TokenStruct, lexer.TokenInterface,
			lexer.TokenImport, lexer.TokenPub, lexer.TokenType, lexer.TokenConst, lexer.TokenAsync:
			return
		}
		p.consume()
	}
}

// syncToNextStatement advances to the next NEWLINE or DEDENT boundary.
func (p *Parser) syncToNextStatement() {
	for {
		raw := p.peekRaw()
		switch raw.Kind {
		case lexer.TokenEOF, lexer.TokenDedent:
			return
		case lexer.TokenNewline:
			p.pos++
			return
		}
		p.pos++
	}
}

// =============================================================================
// Program and top-level declarations
// =============================================================================

func (p *Parser) parseProgram() {
	for !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		decl := p.parseTopLevelDecl()
		if decl != 0 {
			p.tree.AppendChild(0, decl)
		}
		// Safety: ensure progress to prevent infinite loop
		if p.pos == prevPos {
			p.consume()
		}
	}
}

func (p *Parser) parseTopLevelDecl() uint32 {
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenImport:
		return p.parseImportDecl()
	case lexer.TokenPub:
		return p.parsePubDecl()
	case lexer.TokenFn:
		return p.parseFuncDecl(false)
	case lexer.TokenAsync:
		return p.parseAsyncFuncDecl(false)
	case lexer.TokenStruct:
		return p.parseStructDecl(false)
	case lexer.TokenInterface:
		return p.parseInterfaceDecl(false)
	case lexer.TokenType:
		return p.parseTypeAliasDecl(false)
	case lexer.TokenConst:
		return p.parseConstDecl(false)
	default:
		p.errorf(tok, "expected declaration, got %s", tok.Kind)
		p.syncToTopLevel()
		return 0
	}
}

// parsePubDecl handles `pub fn`, `pub async fn`, `pub struct`, `pub interface`,
// `pub type`, `pub const` — called after `pub` is not yet consumed.
func (p *Parser) parsePubDecl() uint32 {
	p.consume() // consume 'pub'
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenFn:
		return p.parseFuncDecl(true)
	case lexer.TokenAsync:
		return p.parseAsyncFuncDecl(true)
	case lexer.TokenPacked, lexer.TokenStruct:
		return p.parseStructDecl(true)
	case lexer.TokenInterface:
		return p.parseInterfaceDecl(true)
	case lexer.TokenType:
		return p.parseTypeAliasDecl(true)
	case lexer.TokenConst:
		return p.parseConstDecl(true)
	default:
		p.errorf(tok, "expected declaration after 'pub', got %s", tok.Kind)
		p.syncToTopLevel()
		return 0
	}
}

// =============================================================================
// Function declarations
// =============================================================================

func (p *Parser) parseFuncDecl(isPub bool) uint32 {
	fnTok, _ := p.expect(lexer.TokenFn)
	node := p.tree.AddNode(ast.NodeFuncDecl, p.tokenIdx(fnTok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}
	p.parseFuncBody(node)
	return node
}

func (p *Parser) parseAsyncFuncDecl(isPub bool) uint32 {
	p.consume() // consume 'async'
	isExtern := p.check(lexer.TokenExtern)
	if isExtern {
		p.consume()
	}
	fnTok, _ := p.expect(lexer.TokenFn)
	node := p.tree.AddNode(ast.NodeFuncDecl, p.tokenIdx(fnTok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}
	p.tree.SetFlags(node, ast.FlagIsAsync)
	if isExtern {
		p.tree.SetFlags(node, ast.FlagIsExtern)
	}
	p.parseFuncBody(node)
	return node
}

// parseFuncBody fills in a NodeFuncDecl node starting from the function name.
func (p *Parser) parseFuncBody(node uint32) {
	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	if p.check(lexer.TokenLBracket) {
		p.tree.SetFlags(node, ast.FlagIsGeneric)
		gp := p.parseGenericParams()
		if gp != 0 {
			p.tree.AppendChild(node, gp)
		}
	}

	p.expect(lexer.TokenLParen)
	for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		param := p.parseParam()
		if param != 0 {
			p.tree.AppendChild(node, param)
		}
		if !p.check(lexer.TokenRParen) {
			p.expect(lexer.TokenComma)
		}
		if p.pos == prevPos {
			p.consume()
		}
	}
	p.expect(lexer.TokenRParen)

	if p.check(lexer.TokenArrow) {
		p.consume()
		retType := p.parseTypeExpr()
		if retType != 0 {
			p.tree.AppendChild(node, retType)
		}
		if p.check(lexer.TokenLBrace) {
			effect := p.parseEffectAnnotation()
			if effect != 0 {
				p.tree.AppendChild(node, effect)
			}
		}
	}

	p.expect(lexer.TokenColon)
	body := p.parseBlock()
	if body != 0 {
		p.tree.AppendChild(node, body)
	}
}

func (p *Parser) parseParam() uint32 {
	tok := p.peek()
	node := p.tree.AddNode(ast.NodeParamDecl, p.tokenIdx(tok))

	switch tok.Kind {
	case lexer.TokenLent:
		p.consume()
		p.tree.SetFlags(node, ast.FlagIsLent)
		tok = p.peek()
	case lexer.TokenBang:
		p.consume()
		p.tree.SetFlags(node, ast.FlagIsSink)
		tok = p.peek()
	case lexer.TokenMut:
		p.consume()
		p.tree.SetFlags(node, ast.FlagIsMut)
		tok = p.peek()
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}
	p.expect(lexer.TokenColon)
	typeNode := p.parseTypeExpr()
	if typeNode != 0 {
		p.tree.AppendChild(node, typeNode)
	}
	return node
}

func (p *Parser) parseGenericParams() uint32 {
	tok := p.peek()
	node := p.tree.AddNode(ast.NodeGenericParams, p.tokenIdx(tok))
	p.expect(lexer.TokenLBracket)
	for !p.check(lexer.TokenRBracket) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		gp := p.parseGenericParam()
		if gp != 0 {
			p.tree.AppendChild(node, gp)
		}
		if !p.check(lexer.TokenRBracket) {
			p.expect(lexer.TokenComma)
		}
		if p.pos == prevPos {
			p.consume()
		}
	}
	p.expect(lexer.TokenRBracket)
	return node
}

func (p *Parser) parseGenericParam() uint32 {
	tok := p.peek()
	node := p.tree.AddNode(ast.NodeGenericParam, p.tokenIdx(tok))
	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}
	if p.check(lexer.TokenColon) {
		p.consume()
		constraint := p.parseTypeExpr()
		if constraint != 0 {
			p.tree.AppendChild(node, constraint)
		}
	}
	return node
}

// parseEffectAnnotation parses `{.raises: [T, ...].}`.
func (p *Parser) parseEffectAnnotation() uint32 {
	tok := p.peek()
	node := p.tree.AddNode(ast.NodeEffectAnnotation, p.tokenIdx(tok))
	p.expect(lexer.TokenLBrace)
	p.expect(lexer.TokenDot)
	p.expect(lexer.TokenIdent) // effect name, e.g. "raises"
	p.expect(lexer.TokenColon)
	p.expect(lexer.TokenLBracket)
	for !p.check(lexer.TokenRBracket) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		t := p.parseTypeExpr()
		if t != 0 {
			p.tree.AppendChild(node, t)
		}
		if !p.check(lexer.TokenRBracket) {
			p.expect(lexer.TokenComma)
		}
		if p.pos == prevPos {
			p.consume()
		}
	}
	p.expect(lexer.TokenRBracket)
	p.expect(lexer.TokenDot)
	p.expect(lexer.TokenRBrace)
	return node
}

// =============================================================================
// Import declarations
// =============================================================================

func (p *Parser) parseImportDecl() uint32 {
	tok, _ := p.expect(lexer.TokenImport)
	node := p.tree.AddNode(ast.NodeImportDecl, p.tokenIdx(tok))

	pathTok, ok := p.expect(lexer.TokenIdent)
	if !ok {
		p.expectNewline()
		return node
	}
	pathID := p.pool.Intern(p.tokenText(pathTok))
	for p.checkRaw(lexer.TokenDot) {
		p.consume()
		seg, ok2 := p.expect(lexer.TokenIdent)
		if !ok2 {
			break
		}
		full := p.pool.Get(pathID) + "." + string(p.tokenText(seg))
		pathID = p.pool.InternString(full)
	}
	p.tree.SetPayload(node, pathID)

	if p.checkRaw(lexer.TokenLBrace) {
		p.consume()
		for !p.check(lexer.TokenRBrace) && !p.check(lexer.TokenEOF) {
			prevPos := p.pos
			nameTok, ok3 := p.expect(lexer.TokenIdent)
			if ok3 {
				nameNode := p.tree.AddNode(ast.NodeIdent, p.tokenIdx(nameTok))
				p.tree.SetPayload(nameNode, p.pool.Intern(p.tokenText(nameTok)))
				p.tree.AppendChild(node, nameNode)
			}
			if !p.check(lexer.TokenRBrace) {
				p.expect(lexer.TokenComma)
			}
			if p.pos == prevPos {
				p.consume()
			}
		}
		p.expect(lexer.TokenRBrace)
	}

	p.expectNewline()
	return node
}

// =============================================================================
// Struct declarations
// =============================================================================

func (p *Parser) parseStructDecl(isPub bool) uint32 {
	isPacked := p.check(lexer.TokenPacked)
	if isPacked {
		p.consume()
	}
	tok, _ := p.expect(lexer.TokenStruct)
	node := p.tree.AddNode(ast.NodeStructDecl, p.tokenIdx(tok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}
	if isPacked {
		p.tree.SetFlags(node, ast.FlagIsPacked)
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	if p.check(lexer.TokenLBracket) {
		p.tree.SetFlags(node, ast.FlagIsGeneric)
		gp := p.parseGenericParams()
		if gp != 0 {
			p.tree.AppendChild(node, gp)
		}
	}

	p.expect(lexer.TokenColon)

	if !p.check(lexer.TokenIndent) {
		p.errorf(p.peek(), "expected INDENT after struct declaration, got %s", p.peek().Kind)
		return node
	}
	p.consume() // consume INDENT

	for !p.check(lexer.TokenDedent) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		inner := p.peek()
		switch inner.Kind {
		case lexer.TokenFn:
			m := p.parseFuncDecl(false)
			if m != 0 {
				p.tree.AppendChild(node, m)
			}
		case lexer.TokenAsync:
			m := p.parseAsyncFuncDecl(false)
			if m != 0 {
				p.tree.AppendChild(node, m)
			}
		case lexer.TokenPub:
			p.consume()
			next := p.peek()
			if next.Kind == lexer.TokenFn {
				m := p.parseFuncDecl(true)
				if m != 0 {
					p.tree.AppendChild(node, m)
				}
			} else if next.Kind == lexer.TokenAsync {
				m := p.parseAsyncFuncDecl(true)
				if m != 0 {
					p.tree.AppendChild(node, m)
				}
			} else {
				f := p.parseFieldDecl(true)
				if f != 0 {
					p.tree.AppendChild(node, f)
				}
			}
		default:
			f := p.parseFieldDecl(false)
			if f != 0 {
				p.tree.AppendChild(node, f)
			}
		}
		if p.pos == prevPos {
			p.consume()
		}
	}

	if p.check(lexer.TokenDedent) {
		p.consume()
	}
	return node
}

func (p *Parser) parseFieldDecl(isPub bool) uint32 {
	isMut := p.check(lexer.TokenMut)
	if isMut {
		p.consume()
	}
	tok := p.peek()
	if tok.Kind != lexer.TokenIdent {
		p.errorf(tok, "expected field name, got %s", tok.Kind)
		p.syncToNextStatement()
		return 0
	}
	p.consume()
	node := p.tree.AddNode(ast.NodeFieldDecl, p.tokenIdx(tok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}
	if isMut {
		p.tree.SetFlags(node, ast.FlagIsMut)
	}
	p.tree.SetPayload(node, p.pool.Intern(p.tokenText(tok)))
	p.expect(lexer.TokenColon)
	typeNode := p.parseTypeExpr()
	if typeNode != 0 {
		p.tree.AppendChild(node, typeNode)
	}
	p.expectNewline()
	return node
}

// =============================================================================
// Interface declarations
// =============================================================================

func (p *Parser) parseInterfaceDecl(isPub bool) uint32 {
	tok, _ := p.expect(lexer.TokenInterface)
	node := p.tree.AddNode(ast.NodeInterfaceDecl, p.tokenIdx(tok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	if p.check(lexer.TokenLBracket) {
		p.tree.SetFlags(node, ast.FlagIsGeneric)
		gp := p.parseGenericParams()
		if gp != 0 {
			p.tree.AppendChild(node, gp)
		}
	}

	p.expect(lexer.TokenColon)

	if !p.check(lexer.TokenIndent) {
		p.errorf(p.peek(), "expected INDENT after interface declaration, got %s", p.peek().Kind)
		return node
	}
	p.consume() // consume INDENT

	for !p.check(lexer.TokenDedent) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		sig := p.parseMethodSig()
		if sig != 0 {
			p.tree.AppendChild(node, sig)
		}
		if p.pos == prevPos {
			p.consume()
		}
	}

	if p.check(lexer.TokenDedent) {
		p.consume()
	}
	return node
}

func (p *Parser) parseMethodSig() uint32 {
	isAsync := p.check(lexer.TokenAsync)
	if isAsync {
		p.consume()
	}
	tok, _ := p.expect(lexer.TokenFn)
	node := p.tree.AddNode(ast.NodeMethodSig, p.tokenIdx(tok))
	if isAsync {
		p.tree.SetFlags(node, ast.FlagIsAsync)
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	if p.check(lexer.TokenLBracket) {
		gp := p.parseGenericParams()
		if gp != 0 {
			p.tree.AppendChild(node, gp)
		}
	}

	p.expect(lexer.TokenLParen)
	for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		param := p.parseParam()
		if param != 0 {
			p.tree.AppendChild(node, param)
		}
		if !p.check(lexer.TokenRParen) {
			p.expect(lexer.TokenComma)
		}
		if p.pos == prevPos {
			p.consume()
		}
	}
	p.expect(lexer.TokenRParen)

	if p.check(lexer.TokenArrow) {
		p.consume()
		ret := p.parseTypeExpr()
		if ret != 0 {
			p.tree.AppendChild(node, ret)
		}
	}
	p.expectNewline()
	return node
}

// =============================================================================
// Type alias and const declarations
// =============================================================================

func (p *Parser) parseTypeAliasDecl(isPub bool) uint32 {
	tok, _ := p.expect(lexer.TokenType)
	node := p.tree.AddNode(ast.NodeTypeAliasDecl, p.tokenIdx(tok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	if p.check(lexer.TokenLBracket) {
		p.tree.SetFlags(node, ast.FlagIsGeneric)
		gp := p.parseGenericParams()
		if gp != 0 {
			p.tree.AppendChild(node, gp)
		}
	}

	p.expect(lexer.TokenEq)

	sumTok := p.peek()
	sumNode := p.tree.AddNode(ast.NodeSumType, p.tokenIdx(sumTok))
	v := p.parseTypeVariant()
	if v != 0 {
		p.tree.AppendChild(sumNode, v)
	}
	// Use checkRaw to avoid consuming the trailing NEWLINE before expectNewline.
	// Sum type variants are always on a single logical line:
	//   type Result = Ok(i32) | Err(string)
	for p.checkRaw(lexer.TokenPipe) {
		p.consume()
		v = p.parseTypeVariant()
		if v != 0 {
			p.tree.AppendChild(sumNode, v)
		}
	}
	p.tree.AppendChild(node, sumNode)

	p.expectNewline()
	return node
}

func (p *Parser) parseTypeVariant() uint32 {
	tok := p.peek()
	if tok.Kind != lexer.TokenIdent {
		return 0
	}
	p.consume()
	node := p.tree.AddNode(ast.NodeVariantDecl, p.tokenIdx(tok))
	p.tree.SetPayload(node, p.pool.Intern(p.tokenText(tok)))

	// Use checkRaw to avoid consuming trailing NEWLINE after unit variants.
	// peek() skips NEWLINEs as a side-effect, which would corrupt p.pos
	// for the caller's expectNewline() call.
	if p.checkRaw(lexer.TokenLParen) {
		p.consume()
		for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
			prevPos := p.pos
			t := p.parseTypeExpr()
			if t != 0 {
				p.tree.AppendChild(node, t)
			}
			if !p.check(lexer.TokenRParen) {
				p.expect(lexer.TokenComma)
			}
			if p.pos == prevPos {
				p.consume()
			}
		}
		p.expect(lexer.TokenRParen)
	}
	return node
}

func (p *Parser) parseConstDecl(isPub bool) uint32 {
	tok, _ := p.expect(lexer.TokenConst)
	node := p.tree.AddNode(ast.NodeConstDecl, p.tokenIdx(tok))
	if isPub {
		p.tree.SetFlags(node, ast.FlagIsPub)
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	p.expect(lexer.TokenColon)
	typeNode := p.parseTypeExpr()
	if typeNode != 0 {
		p.tree.AppendChild(node, typeNode)
	}

	p.expect(lexer.TokenEq)
	expr := p.parseExpr()
	if expr != 0 {
		p.tree.AppendChild(node, expr)
	}
	p.expectNewline()
	return node
}

// =============================================================================
// Block and statement parsing
// =============================================================================

// parseBlock parses INDENT { Stmt } DEDENT.
func (p *Parser) parseBlock() uint32 {
	if !p.check(lexer.TokenIndent) {
		tok := p.peek()
		p.errorf(tok, "expected INDENT, got %s", tok.Kind)
		return 0
	}
	indentTok := p.peek()
	p.consume() // consume INDENT
	node := p.tree.AddNode(ast.NodeBlock, p.tokenIdx(indentTok))

	for !p.check(lexer.TokenDedent) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		stmt := p.parseStmt()
		if stmt != 0 {
			p.tree.AppendChild(node, stmt)
		}
		// Safety: if no progress was made, consume one token to avoid infinite loop
		if p.pos == prevPos {
			p.pos++
		}
	}

	if p.check(lexer.TokenDedent) {
		p.consume()
	}
	return node
}

func (p *Parser) parseStmt() uint32 {
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenLet, lexer.TokenMut:
		return p.parseVarDecl()
	case lexer.TokenImport:
		return p.parseImportDecl()
	case lexer.TokenReturn:
		return p.parseReturnStmt()
	case lexer.TokenIf:
		return p.parseIfStmt()
	case lexer.TokenFor:
		return p.parseForStmt()
	case lexer.TokenWhile:
		return p.parseWhileStmt()
	case lexer.TokenMatch:
		return p.parseMatchStmt()
	case lexer.TokenDefer:
		return p.parseDeferStmt()
	case lexer.TokenUnsafe:
		return p.parseUnsafeBlock()
	case lexer.TokenIn:
		return p.parseArenaBlock()
	default:
		return p.parseAssignOrExprStmt()
	}
}

// parseVarDecl parses `let IDENT [: TypeExpr] = Expr NEWLINE` or `mut ...`.
func (p *Parser) parseVarDecl() uint32 {
	tok := p.consume() // consume 'let' or 'mut'
	node := p.tree.AddNode(ast.NodeVarDecl, p.tokenIdx(tok))
	if tok.Kind == lexer.TokenMut {
		p.tree.SetFlags(node, ast.FlagIsMut)
	}

	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}

	if p.check(lexer.TokenColon) {
		p.consume()
		typeNode := p.parseTypeExpr()
		if typeNode != 0 {
			p.tree.AppendChild(node, typeNode)
		}
	}

	if p.check(lexer.TokenEq) || p.check(lexer.TokenColonEq) {
		p.consume()
		expr := p.parseExpr()
		if expr != 0 {
			p.tree.AppendChild(node, expr)
		}
	}

	p.expectNewline()
	return node
}

func (p *Parser) parseReturnStmt() uint32 {
	tok, _ := p.expect(lexer.TokenReturn)
	node := p.tree.AddNode(ast.NodeReturnStmt, p.tokenIdx(tok))

	raw := p.peekRaw()
	if raw.Kind != lexer.TokenNewline && raw.Kind != lexer.TokenDedent && raw.Kind != lexer.TokenEOF {
		expr := p.parseExpr()
		if expr != 0 {
			p.tree.AppendChild(node, expr)
		}
	}
	p.expectNewline()
	return node
}

func (p *Parser) parseIfStmt() uint32 {
	tok, _ := p.expect(lexer.TokenIf)
	node := p.tree.AddNode(ast.NodeIfStmt, p.tokenIdx(tok))

	cond := p.parseExpr()
	if cond != 0 {
		p.tree.AppendChild(node, cond)
	}
	p.expect(lexer.TokenColon)
	body := p.parseBlock()
	if body != 0 {
		p.tree.AppendChild(node, body)
	}

	for p.check(lexer.TokenElif) {
		elifTok := p.consume()
		elifNode := p.tree.AddNode(ast.NodeElifClause, p.tokenIdx(elifTok))
		elifCond := p.parseExpr()
		if elifCond != 0 {
			p.tree.AppendChild(elifNode, elifCond)
		}
		p.expect(lexer.TokenColon)
		elifBody := p.parseBlock()
		if elifBody != 0 {
			p.tree.AppendChild(elifNode, elifBody)
		}
		p.tree.AppendChild(node, elifNode)
	}

	if p.check(lexer.TokenElse) {
		elseTok := p.consume()
		elseNode := p.tree.AddNode(ast.NodeElseClause, p.tokenIdx(elseTok))
		p.expect(lexer.TokenColon)
		elseBody := p.parseBlock()
		if elseBody != 0 {
			p.tree.AppendChild(elseNode, elseBody)
		}
		p.tree.AppendChild(node, elseNode)
	}

	return node
}

func (p *Parser) parseForStmt() uint32 {
	tok, _ := p.expect(lexer.TokenFor)
	node := p.tree.AddNode(ast.NodeForStmt, p.tokenIdx(tok))

	varTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(varTok)))
	}
	p.expect(lexer.TokenIn)
	iter := p.parseExpr()
	if iter != 0 {
		p.tree.AppendChild(node, iter)
	}
	p.expect(lexer.TokenColon)
	body := p.parseBlock()
	if body != 0 {
		p.tree.AppendChild(node, body)
	}
	return node
}

func (p *Parser) parseWhileStmt() uint32 {
	tok, _ := p.expect(lexer.TokenWhile)
	node := p.tree.AddNode(ast.NodeWhileStmt, p.tokenIdx(tok))

	cond := p.parseExpr()
	if cond != 0 {
		p.tree.AppendChild(node, cond)
	}
	p.expect(lexer.TokenColon)
	body := p.parseBlock()
	if body != 0 {
		p.tree.AppendChild(node, body)
	}
	return node
}

func (p *Parser) parseMatchStmt() uint32 {
	tok, _ := p.expect(lexer.TokenMatch)
	node := p.tree.AddNode(ast.NodeMatchStmt, p.tokenIdx(tok))

	expr := p.parseExpr()
	if expr != 0 {
		p.tree.AppendChild(node, expr)
	}
	p.expect(lexer.TokenColon)

	if !p.check(lexer.TokenIndent) {
		p.errorf(p.peek(), "expected INDENT after match, got %s", p.peek().Kind)
		return node
	}
	p.consume() // consume INDENT

	for !p.check(lexer.TokenDedent) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		arm := p.parseMatchArm()
		if arm != 0 {
			p.tree.AppendChild(node, arm)
		}
		if p.pos == prevPos {
			p.pos++
		}
	}
	if p.check(lexer.TokenDedent) {
		p.consume()
	}
	return node
}

func (p *Parser) parseMatchArm() uint32 {
	tok := p.peek()
	node := p.tree.AddNode(ast.NodeMatchArm, p.tokenIdx(tok))

	pat := p.parsePattern()
	if pat != 0 {
		p.tree.AppendChild(node, pat)
	}
	p.expect(lexer.TokenColon)

	if p.check(lexer.TokenIndent) {
		body := p.parseBlock()
		if body != 0 {
			p.tree.AppendChild(node, body)
		}
	} else {
		expr := p.parseExpr()
		if expr != 0 {
			p.tree.AppendChild(node, expr)
		}
		p.expectNewline()
	}
	return node
}

func (p *Parser) parsePattern() uint32 {
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenIdent:
		p.consume()
		text := p.tokenText(tok)
		// Wildcard: _
		if len(text) == 1 && text[0] == '_' {
			return p.tree.AddNode(ast.NodeWildcardPat, p.tokenIdx(tok))
		}
		// VariantPat with args, or BindingPat
		if p.check(lexer.TokenLParen) {
			node := p.tree.AddNode(ast.NodeVariantPat, p.tokenIdx(tok))
			p.tree.SetPayload(node, p.pool.Intern(text))
			p.consume() // (
			for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
				prevPos := p.pos
				inner := p.parsePattern()
				if inner != 0 {
					p.tree.AppendChild(node, inner)
				}
				if !p.check(lexer.TokenRParen) {
					p.expect(lexer.TokenComma)
				}
				if p.pos == prevPos {
					p.consume()
				}
			}
			p.expect(lexer.TokenRParen)
			return node
		}
		node := p.tree.AddNode(ast.NodeBindingPat, p.tokenIdx(tok))
		p.tree.SetPayload(node, p.pool.Intern(text))
		return node

	case lexer.TokenIntLit, lexer.TokenFloatLit, lexer.TokenStringLit, lexer.TokenCharLit,
		lexer.TokenTrue, lexer.TokenFalse, lexer.TokenNil:
		p.consume()
		return p.tree.AddNode(ast.NodeLiteralPat, p.tokenIdx(tok))

	case lexer.TokenLParen:
		p.consume()
		node := p.tree.AddNode(ast.NodeTuplePat, p.tokenIdx(tok))
		for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
			prevPos := p.pos
			inner := p.parsePattern()
			if inner != 0 {
				p.tree.AppendChild(node, inner)
			}
			if !p.check(lexer.TokenRParen) {
				p.expect(lexer.TokenComma)
			}
			if p.pos == prevPos {
				p.consume()
			}
		}
		p.expect(lexer.TokenRParen)
		return node

	default:
		p.errorf(tok, "expected pattern, got %s", tok.Kind)
		p.consume() // Ensure progress to avoid infinite loop
		return 0
	}
}

func (p *Parser) parseDeferStmt() uint32 {
	tok, _ := p.expect(lexer.TokenDefer)
	node := p.tree.AddNode(ast.NodeDeferStmt, p.tokenIdx(tok))
	expr := p.parseExpr()
	if expr != 0 {
		p.tree.AppendChild(node, expr)
	}
	p.expectNewline()
	return node
}

func (p *Parser) parseUnsafeBlock() uint32 {
	tok, _ := p.expect(lexer.TokenUnsafe)
	node := p.tree.AddNode(ast.NodeUnsafeBlock, p.tokenIdx(tok))
	p.expect(lexer.TokenColon)
	body := p.parseBlock()
	if body != 0 {
		p.tree.AppendChild(node, body)
	}
	return node
}

func (p *Parser) parseArenaBlock() uint32 {
	tok, _ := p.expect(lexer.TokenIn)
	node := p.tree.AddNode(ast.NodeArenaBlock, p.tokenIdx(tok))
	p.expect(lexer.TokenLBracket)
	nameTok, ok := p.expect(lexer.TokenIdent)
	if ok {
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
	}
	p.expect(lexer.TokenRBracket)
	p.expect(lexer.TokenColon)
	body := p.parseBlock()
	if body != 0 {
		p.tree.AppendChild(node, body)
	}
	return node
}

func (p *Parser) parseAssignOrExprStmt() uint32 {
	expr := p.parseExpr()
	if expr == 0 {
		tok := p.peek()
		p.errorf(tok, "expected expression, got %s", tok.Kind)
		p.syncToNextStatement()
		return 0
	}

	tok := p.peekRaw()
	switch tok.Kind {
	case lexer.TokenEq, lexer.TokenPlusEq, lexer.TokenMinusEq,
		lexer.TokenStarEq, lexer.TokenSlashEq, lexer.TokenPercentEq:
		p.consume()
		node := p.tree.AddNode(ast.NodeAssignStmt, p.tokenIdx(tok))
		p.tree.AppendChild(node, expr)
		rhs := p.parseExpr()
		if rhs != 0 {
			p.tree.AppendChild(node, rhs)
		}
		p.expectNewline()
		return node
	default:
		p.expectNewline()
		return expr
	}
}

// =============================================================================
// Type expressions (minimal — full implementation in p03-t05/t06)
// =============================================================================

// parseTypeExpr parses a type expression. Handles the common structural forms
// needed for correct token stream advancement. Full semantics in later tasks.
func (p *Parser) parseTypeExpr() uint32 {
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenIdent, lexer.TokenFuture, lexer.TokenIsolated:
		p.consume()
		node := p.tree.AddNode(ast.NodeTypeExpr, p.tokenIdx(tok))
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(tok)))
		// Handle generic application: Ident[TypeList]
		// Use checkRaw: NEWLINE before '[' means end of type, not a generic.
		if p.checkRaw(lexer.TokenLBracket) {
			genNode := p.tree.AddNode(ast.NodeGenericType, p.tokenIdx(tok))
			p.tree.AppendChild(genNode, node)
			p.consume() // [
			for !p.check(lexer.TokenRBracket) && !p.check(lexer.TokenEOF) {
				prevPos := p.pos
				t := p.parseTypeExpr()
				if t != 0 {
					p.tree.AppendChild(genNode, t)
				}
				if !p.check(lexer.TokenRBracket) {
					p.expect(lexer.TokenComma)
				}
				if p.pos == prevPos {
					p.consume()
				}
			}
			p.expect(lexer.TokenRBracket)
			return genNode
		}
		return node

	case lexer.TokenStar:
		p.consume()
		isMut := p.check(lexer.TokenMut)
		if isMut {
			p.consume()
		}
		node := p.tree.AddNode(ast.NodePtrType, p.tokenIdx(tok))
		if isMut {
			p.tree.SetFlags(node, ast.FlagIsMut)
		}
		inner := p.parseTypeExpr()
		if inner != 0 {
			p.tree.AppendChild(node, inner)
		}
		return node

	case lexer.TokenLBracket:
		p.consume()
		node := p.tree.AddNode(ast.NodeSliceType, p.tokenIdx(tok))
		inner := p.parseTypeExpr()
		if inner != 0 {
			p.tree.AppendChild(node, inner)
		}
		if p.check(lexer.TokenSemicolon) {
			// ArrayType: [T; N]
			p.consume()
			p.tree.Nodes[node].Kind = ast.NodeArrayType
			sizeTok, _ := p.expect(lexer.TokenIntLit)
			sizeNode := p.tree.AddNode(ast.NodeIntLit, p.tokenIdx(sizeTok))
			p.tree.AppendChild(node, sizeNode)
		}
		p.expect(lexer.TokenRBracket)
		return node

	case lexer.TokenFn:
		p.consume()
		node := p.tree.AddNode(ast.NodeFuncType, p.tokenIdx(tok))
		p.expect(lexer.TokenLParen)
		for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
			prevPos := p.pos
			t := p.parseTypeExpr()
			if t != 0 {
				p.tree.AppendChild(node, t)
			}
			if !p.check(lexer.TokenRParen) {
				p.expect(lexer.TokenComma)
			}
			if p.pos == prevPos {
				p.consume()
			}
		}
		p.expect(lexer.TokenRParen)
		if p.check(lexer.TokenArrow) {
			p.consume()
			ret := p.parseTypeExpr()
			if ret != 0 {
				p.tree.AppendChild(node, ret)
			}
		}
		return node

	default:
		p.errorf(tok, "expected type expression, got %s", tok.Kind)
		return 0
	}
}

// =============================================================================
// Expression stub (p03-t05 will replace this with the Pratt parser)
// =============================================================================

// parseExpr parses an expression. This is a minimal stub that handles atoms and
// common postfix operators. The full Pratt parser is implemented in p03-t05.
func (p *Parser) parseExpr() uint32 {
	return p.parseExprWithPrec(bpNone)
}

func (p *Parser) parseExprAtom() uint32 {
	tok := p.peek()
	var node uint32

	switch tok.Kind {
	case lexer.TokenIdent:
		p.consume()
		node = p.tree.AddNode(ast.NodeIdent, p.tokenIdx(tok))
		p.tree.SetPayload(node, p.pool.Intern(p.tokenText(tok)))

	case lexer.TokenIntLit:
		p.consume()
		node = p.tree.AddNode(ast.NodeIntLit, p.tokenIdx(tok))

	case lexer.TokenFloatLit:
		p.consume()
		node = p.tree.AddNode(ast.NodeFloatLit, p.tokenIdx(tok))

	case lexer.TokenStringLit:
		p.consume()
		node = p.tree.AddNode(ast.NodeStringLit, p.tokenIdx(tok))

	case lexer.TokenCharLit:
		p.consume()
		node = p.tree.AddNode(ast.NodeCharLit, p.tokenIdx(tok))

	case lexer.TokenTrue, lexer.TokenFalse:
		p.consume()
		node = p.tree.AddNode(ast.NodeBoolLit, p.tokenIdx(tok))

	case lexer.TokenNil:
		p.consume()
		node = p.tree.AddNode(ast.NodeNilLit, p.tokenIdx(tok))

	case lexer.TokenAwait:
		p.consume()
		awaitNode := p.tree.AddNode(ast.NodeAwaitExpr, p.tokenIdx(tok))
		inner := p.parseExprAtom()
		if inner != 0 {
			p.tree.AppendChild(awaitNode, inner)
		}
		node = awaitNode

	case lexer.TokenSpawn:
		p.consume()
		spawnNode := p.tree.AddNode(ast.NodeSpawnExpr, p.tokenIdx(tok))
		inner := p.parseExprAtom()
		if inner != 0 {
			p.tree.AppendChild(spawnNode, inner)
		}
		node = spawnNode

	default:
		return 0
	}

	// Postfix: call, field access, index, deref.
	// Use peekRaw so NEWLINE terminates the expression rather than being skipped.
	for {
		switch p.peekRaw().Kind {
		case lexer.TokenLParen:
			callTok := p.peek()
			p.consume()
			callNode := p.tree.AddNode(ast.NodeCallExpr, p.tokenIdx(callTok))
			p.tree.AppendChild(callNode, node)
			for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
				prevPos := p.pos
				arg := p.parseExprAtom()
				if arg != 0 {
					p.tree.AppendChild(callNode, arg)
				}
				if !p.check(lexer.TokenRParen) {
					p.expect(lexer.TokenComma)
				}
				if p.pos == prevPos {
					p.consume()
				}
			}
			p.expect(lexer.TokenRParen)
			node = callNode

		case lexer.TokenDot:
			dotTok := p.peek()
			p.consume()
			fieldNode := p.tree.AddNode(ast.NodeFieldExpr, p.tokenIdx(dotTok))
			p.tree.AppendChild(fieldNode, node)
			fieldTok, ok := p.expect(lexer.TokenIdent)
			if ok {
				p.tree.SetPayload(fieldNode, p.pool.Intern(p.tokenText(fieldTok)))
			}
			node = fieldNode

		case lexer.TokenDotStar:
			derefTok := p.peek()
			p.consume()
			derefNode := p.tree.AddNode(ast.NodeDerefExpr, p.tokenIdx(derefTok))
			p.tree.AppendChild(derefNode, node)
			node = derefNode

		case lexer.TokenLBracket:
			idxTok := p.peek()
			p.consume()
			idxNode := p.tree.AddNode(ast.NodeIndexExpr, p.tokenIdx(idxTok))
			p.tree.AppendChild(idxNode, node)
			idx := p.parseExprAtom()
			if idx != 0 {
				p.tree.AppendChild(idxNode, idx)
			}
			p.expect(lexer.TokenRBracket)
			node = idxNode

		default:
			return node
		}
	}
}
