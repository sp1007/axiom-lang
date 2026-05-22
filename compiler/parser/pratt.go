package parser

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

// Binding powers (left binding power for infix, 0 for non-infix)
const (
	bpNone    = 0   // non-operator
	bpOr      = 10  // or
	bpAnd     = 20  // and
	bpNot     = 30  // not (prefix — but used in comparisons too)
	bpCmp     = 40  // == != < > <= >=
	bpBitOr   = 50  // |
	bpBitXor  = 60  // ^
	bpBitAnd  = 70  // &
	bpShift   = 80  // << >>
	bpAdd     = 90  // + -
	bpMul     = 100 // * / %
	bpPower   = 110 // **  (right-associative: LED uses bpPower-1)
	bpUnary   = 120 // unary - ~ not
	bpPostfix = 130 // . .* [] () as spawn await
)

func leftBindingPower(kind lexer.TokenKind) int {
	switch kind {
	case lexer.TokenOr:
		return bpOr
	case lexer.TokenAnd:
		return bpAnd
	case lexer.TokenEqEq, lexer.TokenBangEq,
		lexer.TokenLt, lexer.TokenGt,
		lexer.TokenLtEq, lexer.TokenGtEq:
		return bpCmp
	case lexer.TokenPipe:
		return bpBitOr
	case lexer.TokenCaret:
		return bpBitXor
	case lexer.TokenAmp:
		return bpBitAnd
	case lexer.TokenLtLt, lexer.TokenGtGt:
		return bpShift
	case lexer.TokenPlus, lexer.TokenMinus:
		return bpAdd
	case lexer.TokenStar, lexer.TokenSlash, lexer.TokenPercent:
		return bpMul
	case lexer.TokenStarStar:
		return bpPower
	case lexer.TokenDotDot:
		return 35 // bpRange
	// Postfix
	case lexer.TokenDot, lexer.TokenDotStar,
		lexer.TokenLBracket, lexer.TokenLParen,
		lexer.TokenAs:
		return bpPostfix
	default:
		return bpNone
	}
}

func (p *Parser) parseExprWithPrec(minBP int) uint32 {
	// NUD phase: parse prefix or atom
	left := p.parseNUD()
	if left == 0 {
		return 0
	}

	// LED phase: parse infix/postfix while binding power allows
	for {
		tok := p.peekRaw()
		bp := leftBindingPower(tok.Kind)
		if bp <= minBP {
			break
		}

		left = p.parseLED(left, tok, bp)
		if left == 0 {
			break
		}
	}
	return left
}

func (p *Parser) parseNUD() uint32 {
	tok := p.peek()
	switch tok.Kind {
	case lexer.TokenIdent:
		p.consume()
		node := p.tree.AddNode(ast.NodeIdent, p.tokenIdx(tok))
		nameID := p.pool.Intern(p.tokenText(tok))
		p.tree.SetPayload(node, nameID)
		return node
	case lexer.TokenIntLit:
		p.consume()
		return p.tree.AddNode(ast.NodeIntLit, p.tokenIdx(tok))
	case lexer.TokenFloatLit:
		p.consume()
		return p.tree.AddNode(ast.NodeFloatLit, p.tokenIdx(tok))
	case lexer.TokenStringLit:
		p.consume()
		return p.tree.AddNode(ast.NodeStringLit, p.tokenIdx(tok))
	case lexer.TokenCharLit:
		p.consume()
		return p.tree.AddNode(ast.NodeCharLit, p.tokenIdx(tok))
	case lexer.TokenTrue, lexer.TokenFalse:
		p.consume()
		return p.tree.AddNode(ast.NodeBoolLit, p.tokenIdx(tok))
	case lexer.TokenNil:
		p.consume()
		return p.tree.AddNode(ast.NodeNilLit, p.tokenIdx(tok))
	case lexer.TokenMinus, lexer.TokenTilde, lexer.TokenAmp:
		p.consume()
		node := p.tree.AddNode(ast.NodeUnaryExpr, p.tokenIdx(tok))
		var flags uint16 = 0
		switch tok.Kind {
		case lexer.TokenMinus:
			flags = 1
		case lexer.TokenTilde:
			flags = 3
		case lexer.TokenAmp:
			flags = 4
		}
		p.tree.SetFlags(node, flags)
		operand := p.parseExprWithPrec(bpUnary)
		if operand != 0 {
			p.tree.AppendChild(node, operand)
		}
		return node
	case lexer.TokenNot:
		p.consume()
		node := p.tree.AddNode(ast.NodeUnaryExpr, p.tokenIdx(tok))
		p.tree.SetFlags(node, 2)
		operand := p.parseExprWithPrec(bpNot)
		if operand != 0 {
			p.tree.AppendChild(node, operand)
		}
		return node
	case lexer.TokenLParen:
		return p.parseParenExpr()
	case lexer.TokenLBracket:
		return p.parseArrayLit()
	case lexer.TokenSpawn:
		return p.parseSpawnExpr()
	case lexer.TokenAwait:
		return p.parseAwaitExpr()
	case lexer.TokenPipe:
		return p.parseClosureExpr()
	default:
		p.errorf(tok, "expected expression, got %s", tok.Kind)
		p.consume() // Ensure progress to avoid infinite loops
		return 0
	}
}

func (p *Parser) parseLED(left uint32, opTok lexer.Token, bp int) uint32 {
	p.consume() // consume the operator token
	switch opTok.Kind {
	// Binary operators
	case lexer.TokenOr, lexer.TokenAnd,
		lexer.TokenEqEq, lexer.TokenBangEq,
		lexer.TokenLt, lexer.TokenGt, lexer.TokenLtEq, lexer.TokenGtEq,
		lexer.TokenPipe, lexer.TokenCaret, lexer.TokenAmp,
		lexer.TokenLtLt, lexer.TokenGtGt,
		lexer.TokenPlus, lexer.TokenMinus,
		lexer.TokenStar, lexer.TokenSlash, lexer.TokenPercent,
		lexer.TokenDotDot:
		node := p.tree.AddNode(ast.NodeBinaryExpr, p.tokenIdx(opTok))
		var flags uint16 = 0
		switch opTok.Kind {
		case lexer.TokenEqEq, lexer.TokenBangEq, lexer.TokenLt, lexer.TokenGt, lexer.TokenLtEq, lexer.TokenGtEq:
			flags = 1
		case lexer.TokenAnd, lexer.TokenOr:
			flags = 2
		}
		p.tree.SetFlags(node, flags)
		p.tree.AppendChild(node, left)
		right := p.parseExprWithPrec(bp) // left-associative: same bp
		if right != 0 {
			p.tree.AppendChild(node, right)
		}
		return node
	case lexer.TokenStarStar: // right-associative
		node := p.tree.AddNode(ast.NodeBinaryExpr, p.tokenIdx(opTok))
		p.tree.AppendChild(node, left)
		right := p.parseExprWithPrec(bp - 1) // right-assoc: bp-1
		if right != 0 {
			p.tree.AppendChild(node, right)
		}
		return node
	// Postfix: field access
	case lexer.TokenDot:
		tok := p.peek()
		fieldTok, ok := p.expect(lexer.TokenIdent)
		node := p.tree.AddNode(ast.NodeFieldExpr, p.tokenIdx(tok))
		p.tree.AppendChild(node, left)
		if ok {
			p.tree.SetPayload(node, p.pool.Intern(p.tokenText(fieldTok)))
		}
		fieldName := p.tree.AddNode(ast.NodeIdent, p.tokenIdx(fieldTok))
		p.tree.AppendChild(node, fieldName)
		return node
	case lexer.TokenDotStar: // deref: expr.*
		node := p.tree.AddNode(ast.NodeDerefExpr, p.tokenIdx(opTok))
		p.tree.AppendChild(node, left)
		return node
	// Postfix: index
	case lexer.TokenLBracket:
		node := p.tree.AddNode(ast.NodeIndexExpr, p.tokenIdx(opTok))
		p.tree.AppendChild(node, left)
		idx := p.parseExpr()
		if idx != 0 {
			p.tree.AppendChild(node, idx)
		}
		p.expect(lexer.TokenRBracket)
		return node
	// Postfix: function call
	case lexer.TokenLParen:
		return p.parseCallArgs(left, opTok)
	// Cast: expr as TypeExpr
	case lexer.TokenAs:
		node := p.tree.AddNode(ast.NodeCastExpr, p.tokenIdx(opTok))
		p.tree.AppendChild(node, left)
		typeNode := p.parseTypeExpr()
		if typeNode != 0 {
			p.tree.AppendChild(node, typeNode)
		}
		return node
	default:
		p.errorf(opTok, "unexpected infix operator %s", opTok.Kind)
		return left
	}
}

func (p *Parser) parseCallArgs(callee uint32, lparen lexer.Token) uint32 {
	node := p.tree.AddNode(ast.NodeCallExpr, p.tokenIdx(lparen))
	p.tree.AppendChild(node, callee) // first child is the callee
	for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		// Named arg: ident: expr
		if p.check(lexer.TokenIdent) && p.peekAt(1).Kind == lexer.TokenColon {
			argNode := p.tree.AddNode(ast.NodeNamedArg, p.tokenIdx(p.peek()))
			nameTok := p.consume()
			p.consume() // :
			nameID := p.pool.Intern(p.tokenText(nameTok))
			p.tree.SetPayload(argNode, nameID)
			expr := p.parseExpr()
			if expr != 0 {
				p.tree.AppendChild(argNode, expr)
			}
			p.tree.AppendChild(node, argNode)
		} else {
			expr := p.parseExpr()
			if expr != 0 {
				p.tree.AppendChild(node, expr)
			}
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

func (p *Parser) parseArrayLit() uint32 {
	tok := p.consume() // [
	node := p.tree.AddNode(ast.NodeArrayLit, p.tokenIdx(tok))
	for !p.check(lexer.TokenRBracket) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		elem := p.parseExpr()
		if elem != 0 {
			p.tree.AppendChild(node, elem)
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

func (p *Parser) parseParenExpr() uint32 {
	p.consume() // (
	expr := p.parseExpr()
	p.expect(lexer.TokenRParen)
	return expr // no node wrapping: paren is just for grouping
}

func (p *Parser) parseSpawnExpr() uint32 {
	tok, _ := p.expect(lexer.TokenSpawn)
	node := p.tree.AddNode(ast.NodeSpawnExpr, p.tokenIdx(tok))
	expr := p.parseExprWithPrec(bpUnary)
	if expr != 0 {
		p.tree.AppendChild(node, expr)
	}
	return node
}

func (p *Parser) parseAwaitExpr() uint32 {
	tok, _ := p.expect(lexer.TokenAwait)
	node := p.tree.AddNode(ast.NodeAwaitExpr, p.tokenIdx(tok))
	expr := p.parseExprWithPrec(bpUnary)
	if expr != 0 {
		p.tree.AppendChild(node, expr)
	}
	return node
}

func (p *Parser) parseClosureExpr() uint32 {
	tok, _ := p.expect(lexer.TokenPipe)
	node := p.tree.AddNode(ast.NodeClosureExpr, p.tokenIdx(tok))
	for !p.check(lexer.TokenPipe) && !p.check(lexer.TokenEOF) {
		prevPos := p.pos
		param := p.parseParam()
		if param != 0 {
			p.tree.AppendChild(node, param)
		}
		if !p.check(lexer.TokenPipe) {
			p.expect(lexer.TokenComma)
		}
		if p.pos == prevPos {
			p.consume()
		}
	}
	p.expect(lexer.TokenPipe)
	// Body: block or expression
	if p.check(lexer.TokenColon) {
		body := p.parseBlock()
		if body != 0 {
			p.tree.AppendChild(node, body)
		}
	} else {
		body := p.parseExpr()
		if body != 0 {
			p.tree.AppendChild(node, body)
		}
	}
	return node
}

func (p *Parser) peekAt(offset int) lexer.Token {
	pos := p.pos + offset
	for pos < len(p.tokens) && p.tokens[pos].Kind == lexer.TokenNewline {
		pos++
	}
	if pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.TokenEOF}
	}
	return p.tokens[pos]
}
