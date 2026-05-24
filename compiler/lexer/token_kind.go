package lexer

import "fmt"

// IMPORTANT: TokenKind values must fit in uint8 (max 255).
// TokenKindCount is the sentinel that enforces this in token_kind_test.go.
// Do not reorder existing constants — doing so breaks serialized token streams.
const (
	// Literals
	TokenIntLit    TokenKind = iota // integer literal: 42, 0xFF, 0b1010, 0o77
	TokenFloatLit                   // float literal: 3.14, 1.0e-6
	TokenStringLit                  // string literal: "hello\nworld"
	TokenCharLit                    // char literal: 'a', '\n'

	// Identifier
	TokenIdent // any identifier not matching a keyword

	// Keywords (alphabetical order)
	TokenAnd       // and
	TokenAs        // as
	TokenAsync     // async
	TokenAwait     // await
	TokenConst     // const
	TokenDefer     // defer
	TokenElif      // elif
	TokenElse      // else
	TokenExtern    // extern
	TokenFalse     // false
	TokenFn        // fn
	TokenFor       // for
	TokenFuture    // Future (type-level name)
	TokenIf        // if
	TokenImport    // import
	TokenIn        // in
	TokenInterface // interface
	TokenIsolated  // Isolated (type-level name)
	TokenLent      // lent
	TokenLet       // let
	TokenMatch     // match
	TokenMut       // mut
	TokenNil       // nil
	TokenNot       // not
	TokenOr        // or
	TokenPacked    // packed
	TokenPub       // pub
	TokenReturn    // return
	TokenSpawn     // spawn
	TokenStruct    // struct
	TokenTrue      // true
	TokenType      // type
	TokenUnsafe    // unsafe
	TokenWhile     // while

	// Arithmetic operators
	TokenPlus     // +
	TokenMinus    // -
	TokenStar     // *
	TokenSlash    // /
	TokenPercent  // %
	TokenStarStar // ** (power, right-associative)

	// Comparison operators
	TokenEqEq   // ==
	TokenBangEq // !=
	TokenLt     // <
	TokenGt     // >
	TokenLtEq   // <=
	TokenGtEq   // >=

	// Bitwise operators
	TokenAmp   // & (bitwise and)
	TokenPipe  // | (bitwise or / sum type separator)
	TokenCaret // ^ (bitwise xor)
	TokenTilde // ~ (bitwise not)
	TokenLtLt  // << (left shift)
	TokenGtGt  // >> (right shift)

	// Assignment operators
	TokenEq        // =
	TokenColonEq   // := (declare-and-assign)
	TokenPlusEq    // +=
	TokenMinusEq   // -=
	TokenStarEq    // *=
	TokenSlashEq   // /=
	TokenPercentEq // %=

	// Punctuation
	TokenDot       // .
	TokenDotStar   // .* (deref)
	TokenComma     // ,
	TokenColon     // :
	TokenSemicolon // ;
	TokenArrow     // -> (return type arrow)
	TokenBang      // ! (sink/consume prefix)
	TokenLParen    // (
	TokenRParen    // )
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenLBrace    // {
	TokenRBrace    // }

	// Indentation/structure tokens (synthesized by post-processor, not raw lexer)
	TokenIndent  // increase in indentation level
	TokenDedent  // decrease in indentation level
	TokenNewline // end of logical line

	// Control tokens
	TokenEOF   // end of file
	TokenError // lexer error: bad character or malformed literal

	TokenDotDot // ..

	TokenHash // #

	TokenBreak    // break
	TokenContinue // continue

	TokenKindCount // sentinel — total count; must remain last
)

// String returns a human-readable name for the token kind.
// Used in error messages and debug dumps.
func (k TokenKind) String() string {
	if int(k) < len(tokenKindNames) {
		if s := tokenKindNames[k]; s != "" {
			return s
		}
	}
	return fmt.Sprintf("TokenKind(%d)", k)
}

// tokenKindNames maps each TokenKind to its display name.
var tokenKindNames = [TokenKindCount]string{
	// Literals
	TokenIntLit:    "integer literal",
	TokenFloatLit:  "float literal",
	TokenStringLit: "string literal",
	TokenCharLit:   "char literal",

	// Identifier
	TokenIdent: "identifier",

	// Keywords
	TokenAnd:       "'and'",
	TokenAs:        "'as'",
	TokenAsync:     "'async'",
	TokenAwait:     "'await'",
	TokenConst:     "'const'",
	TokenDefer:     "'defer'",
	TokenElif:      "'elif'",
	TokenElse:      "'else'",
	TokenExtern:    "'extern'",
	TokenFalse:     "'false'",
	TokenFn:        "'fn'",
	TokenFor:       "'for'",
	TokenFuture:    "'Future'",
	TokenIf:        "'if'",
	TokenImport:    "'import'",
	TokenIn:        "'in'",
	TokenInterface: "'interface'",
	TokenIsolated:  "'Isolated'",
	TokenLent:      "'lent'",
	TokenLet:       "'let'",
	TokenMatch:     "'match'",
	TokenMut:       "'mut'",
	TokenNil:       "'nil'",
	TokenNot:       "'not'",
	TokenOr:        "'or'",
	TokenPacked:    "'packed'",
	TokenPub:       "'pub'",
	TokenReturn:    "'return'",
	TokenSpawn:     "'spawn'",
	TokenStruct:    "'struct'",
	TokenTrue:      "'true'",
	TokenType:      "'type'",
	TokenUnsafe:    "'unsafe'",
	TokenWhile:     "'while'",

	// Arithmetic operators
	TokenPlus:     "'+'",
	TokenMinus:    "'-'",
	TokenStar:     "'*'",
	TokenSlash:    "'/'",
	TokenPercent:  "'%'",
	TokenStarStar: "'**'",

	// Comparison operators
	TokenEqEq:   "'=='",
	TokenBangEq: "'!='",
	TokenLt:     "'<'",
	TokenGt:     "'>'",
	TokenLtEq:   "'<='",
	TokenGtEq:   "'>='",

	// Bitwise operators
	TokenAmp:   "'&'",
	TokenPipe:  "'|'",
	TokenCaret: "'^'",
	TokenTilde: "'~'",
	TokenLtLt:  "'<<'",
	TokenGtGt:  "'>>'",

	// Assignment operators
	TokenEq:        "'='",
	TokenColonEq:   "':='",
	TokenPlusEq:    "'+='",
	TokenMinusEq:   "'-='",
	TokenStarEq:    "'*='",
	TokenSlashEq:   "'/='",
	TokenPercentEq: "'%='",

	// Punctuation
	TokenDot:       "'.'",
	TokenDotStar:   "'.*'",
	TokenComma:     "','",
	TokenColon:     "':'",
	TokenSemicolon: "';'",
	TokenArrow:     "'->'",
	TokenBang:      "'!'",
	TokenLParen:    "'('",
	TokenRParen:    "')'",
	TokenLBracket:  "'['",
	TokenRBracket:  "']'",
	TokenLBrace:    "'{'",
	TokenRBrace:    "'}'",

	// Indentation tokens
	TokenIndent:  "INDENT",
	TokenDedent:  "DEDENT",
	TokenNewline: "NEWLINE",

	// Control tokens
	TokenEOF:   "EOF",
	TokenError: "ERROR",
	TokenDotDot:    "'..'",
	TokenHash:      "'#'",
	TokenBreak:     "'break'",
	TokenContinue:  "'continue'",
}

// Keywords maps identifier text to the corresponding keyword TokenKind.
// The lexer uses this after scanning an identifier to determine if the
// identifier is actually a keyword.
var Keywords = map[string]TokenKind{
	"and":       TokenAnd,
	"as":        TokenAs,
	"async":     TokenAsync,
	"await":     TokenAwait,
	"const":     TokenConst,
	"defer":     TokenDefer,
	"elif":      TokenElif,
	"else":      TokenElse,
	"extern":    TokenExtern,
	"false":     TokenFalse,
	"fn":        TokenFn,
	"for":       TokenFor,
	"Future":    TokenFuture,
	"if":        TokenIf,
	"import":    TokenImport,
	"in":        TokenIn,
	"interface": TokenInterface,
	"Isolated":  TokenIsolated,
	"lent":      TokenLent,
	"let":       TokenLet,
	"match":     TokenMatch,
	"mut":       TokenMut,
	"nil":       TokenNil,
	"not":       TokenNot,
	"or":        TokenOr,
	"packed":    TokenPacked,
	"pub":       TokenPub,
	"return":    TokenReturn,
	"spawn":     TokenSpawn,
	"struct":    TokenStruct,
	"true":      TokenTrue,
	"type":      TokenType,
	"unsafe":    TokenUnsafe,
	"while":     TokenWhile,
	"break":     TokenBreak,
	"continue":  TokenContinue,
}

// IsKeyword reports whether kind is a keyword token.
func (k TokenKind) IsKeyword() bool {
	return (k >= TokenAnd && k <= TokenWhile) || k == TokenBreak || k == TokenContinue
}

// IsLiteral reports whether kind is a literal token (int, float, string, char).
func (k TokenKind) IsLiteral() bool {
	return k >= TokenIntLit && k <= TokenCharLit
}

// IsOperator reports whether kind is an operator token.
func (k TokenKind) IsOperator() bool {
	return k >= TokenPlus && k <= TokenPercentEq
}
